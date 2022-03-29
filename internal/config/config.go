package config

import (
	"github.com/kingcobra2468/cot/internal/service"
	"github.com/kingcobra2468/cot/internal/text"
	"github.com/kingcobra2468/cot/internal/text/gvoice"
)

type Services struct {
	Services       []Service `mapstructure:"services"`
	GVoiceNumber   string    `mapstructure:"gvoice_number"`
	TextEncryption bool      `mapstructure:"text_encryption"`
}

type Service struct {
	Name          string   `mapstructure:"name"`
	Domain        string   `mapstructure:"domain"`
	ClientNumbers []string `mapstructure:"client_numbers"`
}

func (s Services) Names() []service.Service {
	services := []service.Service{}
	for _, s := range s.Services {
		services = append(services, service.Service{Name: s.Name, Domain: s.Domain})
	}

	return services
}

func (s Services) Listeners() *[]*text.Listener {
	listeners := []*text.Listener{}
	for _, cs := range s.Services {
		for _, cn := range cs.ClientNumbers {
			if l, err := text.NewListener(gvoice.Link{GVoiceNumber: s.GVoiceNumber, ClientNumber: cn}, s.TextEncryption); err == nil {
				service.AddClient(cs.Name, cn)
				listeners = append(listeners, l)
			}
		}
	}

	return &listeners
}
