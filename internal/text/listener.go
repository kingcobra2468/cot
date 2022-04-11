package text

import (
	"fmt"
	"math"
	"time"

	"github.com/kingcobra2468/cot/internal/config"
	"github.com/kingcobra2468/cot/internal/service"
	"github.com/kingcobra2468/cot/internal/text/crypto"
	"github.com/kingcobra2468/cot/internal/text/gvoice"
	"github.com/kingcobra2468/cot/internal/text/parser"
)

// Listener listens to a given GVoice <-> Client conversation for new commands.
type Listener struct {
	link           gvoice.Link
	latestTextTime uint64
	encryption     bool
}

// minNumMessages is the minimum number of messages to fetch on the first iteration
// when fetching the first conversation chunk.
const minNumMessages uint64 = 5

// Listeners creates a listener for each of the numbers once. Also creates the whitelist
// list for each client number & service pair.
func GenerateListeners(s *config.Services) *[]*Listener {
	listeners := []*Listener{}
	for _, cs := range s.Services {
		for _, cn := range cs.ClientNumbers {
			// check if listener for client number already exists
			if service.ClientExists(cn) {
				service.AddClient(cs.Name, cn)
				continue
			}
			// creates a new client number listener
			if l, err := NewListener(gvoice.Link{GVoiceNumber: s.GVoiceNumber, ClientNumber: cn}, s.TextEncryption); err == nil {
				listeners = append(listeners, l)
				service.AddClient(cs.Name, cn)
			}
		}
	}

	return &listeners
}

// NewListener initializes a new instance of a command listener.
func NewListener(link gvoice.Link, encryption bool) (*Listener, error) {
	currentTime := uint64(time.Now().Unix()) * 1000

	return &Listener{link: link, encryption: encryption,
		latestTextTime: currentTime}, nil
}

// Fetch retrieves the set of new commands that arrived since the last sync.
func (l *Listener) Fetch() *[]service.Command {
	commands := []service.Command{}
	texts, err := l.newTexts()
	if err != nil || len(*texts) == 0 {
		return &commands
	}

	// parses each of the valid texts into a command
	for _, text := range *texts {
		msg := text.Message
		if l.encryption {
			msg, err = crypto.Decrypt(l.link.ClientNumber, msg)
			if err != nil {
				fmt.Println(err)
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

func (l *Listener) SendText(message string) error {
	msg := message
	if l.encryption {
		var err error
		msg, err = crypto.Encrypt(l.link.ClientNumber, msg)
		if err != nil {
			fmt.Println(err)
		}
	}

	fmt.Println(msg)
	return l.link.SendText(msg)
}

// newTexts fetches all of the new text messages since the last sync.
func (l *Listener) newTexts() (*[]gvoice.Text, error) {
	var texts *[]gvoice.Text
	var err error
	// Discovers the set of new texts with the possibility of containing already
	// visited texts. This is done to reduce calls to gvoice and overall api calls.
	for prevSize, multiplier := 0, 1; ; {
		texts, err = l.link.Texts((uint64(prevSize) * uint64(math.Pow(2, float64(multiplier)))) + minNumMessages)
		if err != nil {
			return nil, err
		}
		// check if all possible texts have been retrieved
		currentSize := len(*texts)
		if prevSize == currentSize {
			break
		}

		prevSize = currentSize
	}

	oldestIndex, ok := oldestNewText(texts, l.latestTextTime)
	if !ok {
		return &[]gvoice.Text{}, nil
	}
	// remove all previously executed commands
	prunedTexts := (*texts)[:oldestIndex+1]

	return &prunedTexts, nil
}

// oldestNewText finds the index of the "oldest" unvisited command.
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
