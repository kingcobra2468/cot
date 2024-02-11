package worker

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/kingcobra2468/cot/internal/config"
	"github.com/kingcobra2468/cot/internal/router/worker/crypto"
	"github.com/kingcobra2468/cot/internal/router/worker/gvoice"
	"github.com/kingcobra2468/cot/internal/router/worker/parser"
	"github.com/kingcobra2468/cot/internal/service"
)

// Link binds a given GVoice number with a client number for
// fetching future commands.
type Link struct {
	GVoiceNumber string
	ClientNumber string
}

// Text contains the message and the unix-time timestamp on when
// it was received.
type Text struct {
	Message   string
	Timestamp uint64
}

// GVoiceWorker is a Worker for GVoice.
type GVoiceWorker struct {
	link           Link
	latestTextTime uint64
	encryption     bool
	gvmsClient     gvoice.GVoiceClient
}

// minNumMessages is the minimum number of messages to fetch on the first iteration
// when fetching the first conversation chunk.
const minNumMessages uint64 = 5

// GenerateGVoiceWorkers creates a list of Worker instances from the configuration file. This
// will also add each of the client numbers to the whitelist in the process.
func GenerateGVoiceWorkers(c *config.Services, gvc gvoice.GVoiceClient) *[]*GVoiceWorker {
	workers := []*GVoiceWorker{NewGVoiceWorker(Link{GVoiceNumber: c.GVoiceNumber, ClientNumber: c.GVoiceNumber}, false, gvc)}
	for _, s := range c.Services {
		for _, cn := range s.ClientNumbers {
			// check if worker exists (to avoid duplicate workers)
			if service.ClientExists(cn) {
				service.AddClient(s.Name, cn)
				continue
			}

			workers = append(workers, NewGVoiceWorker(Link{GVoiceNumber: c.GVoiceNumber, ClientNumber: cn}, c.TextEncryption, gvc))
			service.AddClient(s.Name, cn)
			glog.Infof("created new gvoice worker for %s", cn)
		}
	}

	return &workers
}

// NewGVoiceWorker creates a new instance of GVoice source worker.
func NewGVoiceWorker(link Link, encryption bool, c gvoice.GVoiceClient) *GVoiceWorker {
	// get the current time to prevent old commands (those which existed prior to start of cot)
	// from being executed
	currentTime := uint64(time.Now().Unix()) * 1000

	return &GVoiceWorker{link: link, encryption: encryption,
		latestTextTime: currentTime, gvmsClient: c}
}

// Fetch retrieves the set of new commands since the last sync.
func (gw *GVoiceWorker) Fetch() *[]service.UserInput {
	commands := []service.UserInput{}
	texts, err := gw.unprocessedTexts()
	if err != nil || len(*texts) == 0 {
		return &commands
	}

	// parses each of the valid texts into a command
	for _, text := range *texts {
		msg := text.Message
		// perform decryption of message if enabled
		if gw.encryption {
			msg, err = crypto.Decrypt(gw.link.ClientNumber, msg)
			if err != nil {
				glog.Errorln(err)
				continue
			}
		}

		if command, err := parser.Parse(msg); err == nil {
			commands = append(commands, *command)
		}
	}
	// update the timestamp to that of the last recorded command
	gw.latestTextTime = (*texts)[len(*texts)-1].Timestamp

	return &commands
}

// unprocessedTexts fetches all of the new text messages since the last sync.
func (gw *GVoiceWorker) unprocessedTexts() (*[]Text, error) {
	var texts *[]Text
	var err error
	// Discovers the set of new texts with the possibility of containing already
	// visited texts. This is done to reduce calls to gvoice and overall api calls.
	for prevSize, multiplier := 0, 1; ; {
		// increase number of messages to search for by following the sequence 2^n
		texts, err = gw.newTexts((uint64(prevSize) * uint64(math.Pow(2, float64(multiplier)))) + minNumMessages)
		if err != nil {
			return nil, err
		}
		// check if all possible texts have been retrieved
		currentSize := len(*texts)
		if prevSize == currentSize || (prevSize > 0 && (*texts)[len(*texts)-1].Timestamp < gw.latestTextTime) {
			break
		}

		prevSize = currentSize
	}

	// fetch the index of "oldest" newest (message that is yet to be executed) in order
	// to prune the already executed messages from the list of messages
	oldestIndex, ok := oldestNewText(texts, gw.latestTextTime)
	if !ok {
		return &[]Text{}, nil
	}
	// remove all previously executed commands
	prunedTexts := (*texts)[:oldestIndex+1]

	return &prunedTexts, nil
}

// Texts fetches the number of messages specified from the message history between
// a client number and an associated gvoice number.
func (gw *GVoiceWorker) newTexts(numMessages uint64) (*[]Text, error) {
	texts := []Text{}
	// extract a list of messages which contain at most numMessages messages
	msgList, err := gw.gvmsClient.GetContactHistory(context.Background(),
		&gvoice.FetchContactHistoryRequest{GvoicePhoneNumber: &gw.link.GVoiceNumber,
			RecipientPhoneNumber: &gw.link.ClientNumber, NumMessages: &numMessages})

	if err != nil {
		return &texts, err
	}
	if !*msgList.Success {
		return &texts, errors.New(*msgList.Error)
	}

	// creates a Text instance for each raw message
	for _, text := range msgList.Messages {
		if !*text.Source {
			continue
		}
		texts = append(texts, Text{Message: *text.MessageContents, Timestamp: uint64(*text.Timestamp)})
	}

	return &texts, nil
}

// oldestNewText finds the index of the "oldest" unvisited command. Commands arrive in
// newest to oldest order by nature of GVoice API.
func oldestNewText(texts *[]Text, timestamp uint64) (int, bool) {
	oldestIndex := len(*texts) - 1
	newTextFound := false
	for i := range *texts {
		if (*texts)[oldestIndex-i].Timestamp > timestamp {
			oldestIndex = oldestIndex - i
			newTextFound = true
			break
		}
	}

	return oldestIndex, newTextFound
}

// Send sends a text message to the recipient.
func (gw *GVoiceWorker) Send(message string) error {
	msg := encode(message)
	if gw.encryption {
		var err error
		msg, err = crypto.Encrypt(gw.link.ClientNumber, msg)
		if err != nil {
			glog.Errorln(err)
		}
	}

	req := &gvoice.SendSMSRequest{GvoicePhoneNumber: &gw.link.GVoiceNumber, RecipientPhoneNumber: &gw.link.ClientNumber, Message: &msg}
	resp, err := gw.gvmsClient.SendSMS(context.Background(), req)

	if err != nil {
		return err
	}
	if resp.Error != nil {
		return errors.New(*resp.Error)
	}

	return nil
}

// encode encodes a string for GVoice.
func encode(message string) string {
	// undo any existing encoding on quotes
	message = strings.ReplaceAll(message, "\\\"", "\"")
	// encode all quotes
	message = strings.ReplaceAll(message, "\"", "\\\"")
	// remove all newlines as they cannot exist when sending messages with gvms
	message = strings.ReplaceAll(message, "\n", "")

	return message
}

func (gw *GVoiceWorker) LoopBack() bool {
	return strings.EqualFold(gw.link.GVoiceNumber, gw.link.ClientNumber)
}

func (gw *GVoiceWorker) Recipient() string {
	return gw.link.ClientNumber
}
