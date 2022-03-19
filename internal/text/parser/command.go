package parser

import (
	"errors"
	"strings"

	"github.com/kingcobra2468/cot/internal/service"
)

type CommandParser struct {
	Encryption    bool
	Authorization bool
}

var (
	errUnparsableCommand = errors.New("unable to parse command")
)

// Parse parses the input text into command name and argument components.
// TODO: implement logic for encryption and encryption in this function
func (cp CommandParser) Parse(text string) (*service.Command, error) {
	tokens := strings.Fields(text)

	if len(tokens) == 0 {
		return nil, errUnparsableCommand
	}

	return &service.Command{Name: tokens[0], Arguments: tokens[1:]}, nil
}
