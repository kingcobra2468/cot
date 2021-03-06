// parser handles the parsing of a text into viable command components.
// Also manages encryption if enabled.
package parser

import (
	"errors"
	"strings"

	"github.com/kingcobra2468/cot/internal/service"
)

var (
	errUnparsableCommand = errors.New("unable to parse command")
)

// Parse parses the input text into an instance of a Command.
func Parse(text string) (*service.UserInput, error) {
	tokens := strings.Fields(text)
	if len(tokens) == 0 {
		return nil, errUnparsableCommand
	}

	return &service.UserInput{Name: strings.ToLower(tokens[0]), Args: tokens[1:], Raw: text}, nil
}
