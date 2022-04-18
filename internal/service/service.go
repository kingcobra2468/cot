package service

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/kingcobra2468/cot/internal/config"
)

// Service handles the communication between a command request and the associated
// client service.
type Service struct {
	Name     string
	BaseURI  string
	Endpoint string
}

// Command sets up the schema for a command request via the name of a command and its
// its arguments.
type Command struct {
	Name      string
	Arguments []string
}

// CommandRequest contains the JSON request schema.
type CommandRequest struct {
	Args []string `json:"args"`
}

// CommandRequest sets up the JSON response schema.
type CommandResponse struct {
	Message string `json:"message"`
	Error   error  `json:"error,omitempty"`
}

// GenerateServices creates a list of services that were specified
// in the configuration file.
func GenerateServices(c *config.Services) []Service {
	services := []Service{}
	for _, s := range c.Services {
		services = append(services, Service{Name: s.Name, BaseURI: s.BaseURI, Endpoint: s.Endpoint})
	}

	return services
}

// Execute will push the command request to the associated client service and will
// retrieve the output.
func (s Service) Execute(c *Command) (string, error) {
	client := &http.Client{Timeout: time.Second * 10}
	data, err := json.Marshal(CommandRequest{Args: c.Arguments})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", s.BaseURI+s.Endpoint, bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var output CommandResponse
	err = json.Unmarshal(bodyBytes, &output)
	// check if output parsable
	if err != nil {
		return "", err
	}
	// check if error was sent back from client service
	if output.Error != nil {
		return "", err
	}

	return output.Message, err
}
