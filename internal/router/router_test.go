package router

import (
	"sync"
	"testing"
	"time"

	"github.com/kingcobra2468/cot/internal/router/mocks"
	"github.com/kingcobra2468/cot/internal/service"
	"github.com/kingcobra2468/cot/internal/testutil"
	"github.com/stretchr/testify/mock"
)

const commandName = "test"
const recipientNumber = "1"
const coolDown = time.Duration(time.Second * 3)

func stopEventLoop(t *testing.T, done chan struct{}, timeout time.Duration) {
	time.Sleep(time.Duration(timeout))
	t.Log("stopped event loop")
	done <- struct{}{}
}

func TestProcess(t *testing.T) {
	server := testutil.NewStubServer(t)
	s := testutil.NewFakeService(server, commandName, recipientNumber)

	cache := *service.NewCache()
	cache.Add(*s)

	mockWorker := mocks.NewWorker(t)
	mockWorker.EXPECT().Fetch().Return(&[]service.UserInput{{Name: commandName, Raw: "test"}})
	mockWorker.On("LoopBack").Return(false)
	mockWorker.On("Recipient").Return(recipientNumber)
	mockWorker.On("Send", mock.Anything).Return(nil)

	el := NewEventLoop(2, 2, coolDown, &cache)
	el.AddWorker(mockWorker)

	done := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(1)

	go stopEventLoop(t, done, time.Second*11)

	t.Log("started event loop")
	el.Start(done, &wg)

	wg.Wait()

	mockWorker.AssertExpectations(t)
	mockWorker.AssertCalled(t, "Fetch", mock.Anything)
	mockWorker.AssertCalled(t, "Send", mock.Anything)
}

func TestProcess_stress(t *testing.T) {
	server := testutil.NewStubServer(t)
	s := testutil.NewFakeService(server, commandName, recipientNumber)

	cache := *service.NewCache()
	cache.Add(*s)

	mockWorkers := make([]*mocks.Worker, 8)
	el := NewEventLoop(8, 6, coolDown, &cache)

	for i := 0; i < 8; i++ {
		mockWorker := mocks.NewWorker(t)
		mockWorker.EXPECT().Fetch().Return(&[]service.UserInput{{Name: commandName, Raw: "test"}})
		mockWorker.On("LoopBack").Return(false)
		mockWorker.On("Recipient").Return(recipientNumber)
		mockWorker.On("Send", mock.Anything).Return(nil)

		//mockWorkers = append(mockWorkers, mockWorker)
		mockWorkers[i] = mockWorker
		el.AddWorker(mockWorker)
	}

	done := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(1)

	go stopEventLoop(t, done, time.Second*20)

	t.Log("started event loop")
	el.Start(done, &wg)

	wg.Wait()

	for _, mockWorker := range mockWorkers {
		mockWorker.AssertExpectations(t)
		mockWorker.AssertCalled(t, "Fetch", mock.Anything)
		mockWorker.AssertCalled(t, "Send", mock.Anything)
	}
}

func TestProcess_ping(t *testing.T) {
	server := testutil.NewStubServer(t)
	s := testutil.NewFakeService(server, commandName, recipientNumber)

	cache := *service.NewCache()
	cache.Add(*s)

	mockWorker := mocks.NewWorker(t)
	mockWorker.EXPECT().Fetch().Return(&[]service.UserInput{{Name: "ping", Raw: "ping"}})
	mockWorker.EXPECT().Send("pong").Return(nil)
	mockWorker.On("Send", mock.Anything).Return(nil)

	el := NewEventLoop(2, 2, coolDown, &cache)
	el.AddWorker(mockWorker)

	done := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(1)

	go stopEventLoop(t, done, time.Second*11)

	t.Log("started event loop")
	el.Start(done, &wg)

	wg.Wait()

	mockWorker.AssertExpectations(t)
}

func TestProcess_pong(t *testing.T) {
	server := testutil.NewStubServer(t)
	s := testutil.NewFakeService(server, commandName, recipientNumber)

	cache := *service.NewCache()
	cache.Add(*s)

	mockWorker := mocks.NewWorker(t)
	mockWorker.EXPECT().Fetch().Return(&[]service.UserInput{{Name: "pong", Raw: "pong"}})
	mockWorker.On("LoopBack").Return(true)

	el := NewEventLoop(2, 2, coolDown, &cache)
	el.AddWorker(mockWorker)

	done := make(chan struct{})
	wg := sync.WaitGroup{}
	wg.Add(1)

	go stopEventLoop(t, done, time.Second*11)

	t.Log("started event loop")
	el.Start(done, &wg)

	wg.Wait()

	mockWorker.AssertExpectations(t)
}
