package router

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/kingcobra2468/cot/internal/router/mocks"
	"github.com/kingcobra2468/cot/internal/service"
	"github.com/stretchr/testify/mock"
)

const commandName = "test"
const recipientNumber = "1"

func NewStubServer(t *testing.T) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/test" {
			t.Errorf("Expected to request '/test', got: %s", r.URL.Path)
		}
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("Expected Accept: application/json header, got: %s", r.Header.Get("Accept"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"value":"fixed"}`))
	}))
	t.Log("started service test server")

	t.Cleanup(func() {
		server.Close()
	})

	return server
}

func NewFakeService(server *httptest.Server) *service.Service {
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

func TestProcess(t *testing.T) {
	server := NewStubServer(t)
	s := NewFakeService(server)

	cache := *service.NewCache()
	cache.Add(*s)

	mockWorker := mocks.NewWorker(t)
	mockWorker.EXPECT().Fetch().Return(&[]service.UserInput{service.UserInput{Name: commandName, Raw: "test"}})
	mockWorker.On("LoopBack").Return(false)
	mockWorker.On("Recipient").Return(recipientNumber)
	mockWorker.On("Send", mock.Anything).Return(nil)

	el := NewEventLoop(2, 2, &cache)
	el.AddWorker(mockWorker)

	done := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func(done chan struct{}) {
		time.Sleep(time.Duration(time.Second * 11))
		t.Log("stopped event loop")
		done <- struct{}{}
	}(done)

	t.Log("start event loop")
	el.Start(done, &wg)

	wg.Wait()

	mockWorker.AssertExpectations(t)
	mockWorker.AssertCalled(t, "Fetch", mock.Anything)
	mockWorker.AssertCalled(t, "Send", mock.Anything)
}

func TestProcess_ping(t *testing.T) {
	server := NewStubServer(t)
	s := NewFakeService(server)

	cache := *service.NewCache()
	cache.Add(*s)

	mockWorker := mocks.NewWorker(t)
	mockWorker.EXPECT().Fetch().Return(&[]service.UserInput{service.UserInput{Name: "ping", Raw: "ping"}})
	mockWorker.EXPECT().Send("pong").Return(nil)
	mockWorker.On("Send", mock.Anything).Return(nil)

	el := NewEventLoop(2, 2, &cache)
	el.AddWorker(mockWorker)

	done := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func(done chan struct{}) {
		time.Sleep(time.Duration(time.Second * 11))
		t.Log("stopped event loop")
		done <- struct{}{}
	}(done)

	t.Log("start event loop")
	el.Start(done, &wg)

	wg.Wait()

	mockWorker.AssertExpectations(t)
}

func TestProcess_pong(t *testing.T) {
	server := NewStubServer(t)
	s := NewFakeService(server)

	cache := *service.NewCache()
	cache.Add(*s)

	mockWorker := mocks.NewWorker(t)
	mockWorker.EXPECT().Fetch().Return(&[]service.UserInput{service.UserInput{Name: "pong", Raw: "pong"}})
	mockWorker.On("LoopBack").Return(true)

	el := NewEventLoop(2, 2, &cache)
	el.AddWorker(mockWorker)

	done := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func(done chan struct{}) {
		time.Sleep(time.Duration(time.Second * 11))
		t.Log("stopped event loop")
		done <- struct{}{}
	}(done)

	t.Log("start event loop")
	el.Start(done, &wg)

	wg.Wait()

	mockWorker.AssertExpectations(t)
}
