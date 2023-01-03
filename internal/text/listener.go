package text

import (
	"math"
	"time"

	"github.com/golang/glog"
	"github.com/kingcobra2468/cot/internal/config"
	"github.com/kingcobra2468/cot/internal/service"
	"github.com/kingcobra2468/cot/internal/text/crypto"
	"github.com/kingcobra2468/cot/internal/text/gvoice"
	"github.com/kingcobra2468/cot/internal/text/parser"
)

// Listener listens to a given GVoice <-> Client conversation for new commands.
type Listener struct {
	link            gvoice.Link
	latestTextTime  uint64
	encryption      bool
	timestampOffset uint64
}

// minNumMessages is the minimum number of messages to fetch on the first iteration
// when fetching the first conversation chunk.
const minNumMessages uint64 = 5
const pingOffset uint64 = 5

// GenerateListeners creates a list of Listener instances from the configuration file. This
// will also add each of the client numbers to the whitelist in the process.
func GenerateListeners(c *config.Services) *[]*Listener {
	listeners := []*Listener{NewListener(gvoice.Link{GVoiceNumber: c.GVoiceNumber, ClientNumber: c.GVoiceNumber}, false, pingOffset)}
	for _, s := range c.Services {
		for _, cn := range s.ClientNumbers {
			// check if listener for client number already exists
			if service.ClientExists(cn) {
				service.AddClient(s.Name, cn)
				continue
			}
			// creates a new client number listener
			listeners = append(listeners, NewListener(gvoice.Link{GVoiceNumber: c.GVoiceNumber, ClientNumber: cn}, c.TextEncryption, 0))
			service.AddClient(s.Name, cn)
			glog.Infof("created new listener for %s", cn)
		}
	}

	return &listeners
}

// NewListener initializes a new instance of a command listener.
func NewListener(link gvoice.Link, encryption bool, timestampOffset uint64) *Listener {
	// get the current time to prevent old commands (those which existed prior to start of cot)
	// from being executed
	currentTime := uint64(time.Now().Unix()) * 1000

	return &Listener{link: link, encryption: encryption,
		latestTextTime: currentTime, timestampOffset: timestampOffset}
}

// Fetch retrieves the set of new commands that arrived since the last sync.
func (l *Listener) Fetch() *[]service.UserInput {
	commands := []service.UserInput{}
	texts, err := l.newTexts()
	if err != nil || len(*texts) == 0 {
		return &commands
	}

	// parses each of the valid texts into a command
	for _, text := range *texts {
		msg := text.Message
		// perform decryption of message if enabled
		if l.encryption {
			msg, err = crypto.Decrypt(l.link.ClientNumber, msg)
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
	l.latestTextTime = (*texts)[len(*texts)-1].Timestamp

	return &commands
}

// SendText sends a text message to the recipient in the link.
func (l *Listener) SendText(message string) error {
	msg := message
	// perform encryption of message if enabled
	if l.encryption {
		var err error
		msg, err = crypto.Encrypt(l.link.ClientNumber, msg)
		if err != nil {
			glog.Errorln(err)
		}
	}

	return l.link.SendText(msg)
}

// newTexts fetches all of the new text messages since the last sync.
func (l *Listener) newTexts() (*[]gvoice.Text, error) {
	var texts *[]gvoice.Text
	var err error
	// Discovers the set of new texts with the possibility of containing already
	// visited texts. This is done to reduce calls to gvoice and overall api calls.
	for prevSize, multiplier := 0, 1; ; {
		// increase number of messages to search for by following the sequence 2^n
		texts, err = l.link.Texts((uint64(prevSize) * uint64(math.Pow(2, float64(multiplier)))) + minNumMessages)
		if err != nil {
			return nil, err
		}
		// check if all possible texts have been retrieved
		currentSize := len(*texts)
		if prevSize == currentSize || (prevSize > 0 && (*texts)[len(*texts)-1].Timestamp < l.latestTextTime) {
			break
		}

		prevSize = currentSize
	}

	// fetch the index of "oldest" newest (message that is yet to be executed) in order
	// to prune the already executed messages from the list of messages
	oldestIndex, ok := oldestNewText(texts, l.latestTextTime)
	if !ok {
		return &[]gvoice.Text{}, nil
	}
	// remove all previously executed commands
	prunedTexts := (*texts)[:oldestIndex+1]

	return &prunedTexts, nil
}

// oldestNewText finds the index of the "oldest" unvisited command. Commands arrive in
// newest to oldest order by nature of GVoice API.
func oldestNewText(texts *[]gvoice.Text, timestamp uint64) (int, bool) {
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
