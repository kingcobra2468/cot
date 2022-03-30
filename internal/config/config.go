package config

import (
	"github.com/kingcobra2468/cot/internal/service"
	"github.com/kingcobra2468/cot/internal/text"
	"github.com/kingcobra2468/cot/internal/text/gvoice"
)

// Services contains configuration on each of the services and the client
// numbers authorized to use it. Also contains the encryption and gvoice number
//  bindings that the client numbers send messages to.
type Services struct {
	Services       []Service `mapstructure:"services"`
	GVoiceNumber   string    `mapstructure:"gvoice_number"`
	TextEncryption bool      `mapstructure:"text_encryption"`
}

// GVMSConfig contains configuration on communicating with GVMS.
type GVMSConfig struct {
	Hostname string `mapstructure:"hostname"`
	Port     int    `mapstructure:"port"`
}

// Service contains configuration on the service name (also used as the command name)
// as well as the service domain. Also contains a list of client numbers authorized
// to use this service.
type Service struct {
	Name          string   `mapstructure:"name"`
	Domain        string   `mapstructure:"domain"`
	ClientNumbers []string `mapstructure:"client_numbers"`
}

// Names returns a list of all of the service names.
func (s Services) Names() []service.Service {
	services := []service.Service{}
	for _, s := range s.Services {
		services = append(services, service.Service{Name: s.Name, Domain: s.Domain})
	}

	return services
}

// Listeners creates a listener for each of the numbers once. Also creates the whitelist
// list for each client number & service pair.
func (s Services) Listeners() *[]*text.Listener {
	listeners := []*text.Listener{}
	for _, cs := range s.Services {
		for _, cn := range cs.ClientNumbers {
			// check if listener for client number already exists
			if service.ClientExists(cn) {
				service.AddClient(cs.Name, cn)
				continue
			}
			// creates a new client number listener
			if l, err := text.NewListener(gvoice.Link{GVoiceNumber: s.GVoiceNumber, ClientNumber: cn}, s.TextEncryption); err == nil {
				listeners = append(listeners, l)
				service.AddClient(cs.Name, cn)
			}
		}
	}

	return &listeners
}
