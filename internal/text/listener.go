package text

import (
	"math"
	"time"

	"github.com/kingcobra2468/cot/internal/service"
	"github.com/kingcobra2468/cot/internal/text/gvoice"
	"github.com/kingcobra2468/cot/internal/text/parser"
)

// Listener listens to a given GVoice <-> Client conversation for new commands.
type Listener struct {
	link           gvoice.Link
	latestTextTime uint64
	parser         parser.CommandParser
}

// minNumMessages is the minimum number of messages to fetch on the first iteration
// when fetching the first conversation chunk.
const minNumMessages uint64 = 5

// NewListener initializes a new instance of a command listener.
func NewListener(link gvoice.Link, encryption bool) (*Listener, error) {
	currentTime := uint64(time.Now().Unix()) * 1000

	return &Listener{link: link, parser: parser.CommandParser{Encryption: encryption, Authorization: true},
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
		if command, err := l.parser.Parse(text.Message); err == nil {
			commands = append(commands, *command)
		}
	}
	// update the timestamp to that of the last recorded command
	l.latestTextTime = (*texts)[len(*texts)-1].Timestamp

	return &commands
}

func (l *Listener) SendText(message string) error {
	return l.link.SendText(message)
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
