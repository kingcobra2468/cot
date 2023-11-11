package testutil

import (
	"net/http/httptest"

	"github.com/kingcobra2468/cot/internal/service"
)

func NewFakeService(server *httptest.Server, commandName string, recipientNumber string) *service.Service {
	c := service.Command{}
	c.Endpoint = "/test"
	c.Method = "get"
	c.Response.Type = service.PlainTextResponse

	s := service.Service{}
	s.BaseURI = server.URL
	s.Name = commandName
	s.Meta = map[string]*service.Command{commandName: &c}

	service.AddClient(commandName, recipientNumber)

	return &s
}
