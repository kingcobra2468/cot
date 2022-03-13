package text

import (
	"github.com/kingcobra2468/cot/internal/service"
	"github.com/kingcobra2468/cot/internal/text/gvoice"
)

// Listener listens to a given GVoice <-> Client conversation for new commands.
type Listener struct {
	gvoice.Link
	Encryption     bool
	latestTextHash string
	latestTextTime int
}

// Fetch retrieves the set of new commands that arrived since the last synced timestamp.
func (l *Listener) Fetch() *[]service.Command {
	return &[]service.Command{{Name: l.ClientNumber, Arguments: []string{l.Link.ClientNumber, l.Link.GVoiceNumber}}}
}
