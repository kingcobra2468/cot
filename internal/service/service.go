package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/kingcobra2468/cot/internal/config"
)

// Service pushes a given command to its intended service.
type Service struct {
	Name    string
	BaseURI string
}

// Command sets up the schema for a given command.
type Command struct {
	Name      string
	Arguments []string
}

// CommandRequest setsup up the request payload for sending command arguments to
// the service client.
type CommandRequest struct {
	Args []string `json:"args"`
}

// CommandRequest setsup up the response payload for sending command arguments to
// the service client
type CommandResponse struct {
	Message string `json:"message"`
	Error   error  `json:"error,omitempty"`
}

// Names returns a list of all of the service names.
func GenerateServices(s *config.Services) []Service {
	services := []Service{}
	for _, s := range s.Services {
		services = append(services, Service{Name: s.Name, BaseURI: s.BaseURI})
	}

	return services
}

// Listen attends to a given channel for new commands and executes
// them against a given service.
func (s Service) Listen(stream <-chan *Command, done <-chan struct{}) {
	go func() {
		for {
			select {
			case <-done:
				return
			case command := <-stream:
				message, err := s.Execute(command)
				fmt.Println(message, err)
			}
		}
	}()
}

// execute will run the command against the service.
func (s Service) Execute(c *Command) (string, error) {
	client := &http.Client{Timeout: time.Second * 10}
	data, err := json.Marshal(CommandRequest{Args: c.Arguments})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/cmd", s.BaseURI), bytes.NewBuffer(data))
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
	if err != nil {
		return "", err
	}
	if output.Error != nil {
		return "", err
	}

	return output.Message, err
}
