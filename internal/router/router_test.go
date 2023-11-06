package router

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/kingcobra2468/cot/internal/router"
	"github.com/kingcobra2468/cot/internal/router/mocks"
	"github.com/kingcobra2468/cot/internal/service"
	"github.com/stretchr/testify/mock"
)

func TestProcess(t *testing.T) {
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

	defer server.Close()

	c := service.Command{}
	c.Endpoint = "/test"
	c.Method = "get"
	c.Response.Type = service.PlainTextResponse

	cs := map[string]*service.Command{"test": &c}

	s := service.Service{}
	s.BaseURI = server.URL
	s.Name = "test"
	s.Meta = cs

	cache := *service.NewCache()
	cache.Add(s)

	service.AddClient("test", "t")

	mockWorker := mocks.NewWorker(t)
	mockWorker.EXPECT().Fetch().Return(&[]service.UserInput{service.UserInput{Name: "test", Raw: "test"}})
	mockWorker.On("LoopBack").Return(false)
	mockWorker.On("Recipient").Return("t")
	mockWorker.On("Send", mock.Anything).Return(nil)

	el := router.NewEventLoop(2, 2, &cache)
	el.AddWorker(mockWorker)

	done := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func(done chan struct{}) {
		time.Sleep(time.Duration(time.Second * 11))
		t.Log("stopping eventloop")
		done <- struct{}{}
	}(done)

	t.Log("start eventloop")
	el.Start(done, &wg)

	wg.Wait()

	mockWorker.AssertExpectations(t)
	mockWorker.AssertCalled(t, "Fetch", mock.Anything)
	mockWorker.AssertCalled(t, "Send", mock.Anything)
}
