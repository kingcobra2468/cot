// text handles the stream of commands from client numbers and the overall
// runtime.
package router

import (
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/kingcobra2468/cot/internal/service"
)

// EventLoop handles new command requests received by workers.
type EventLoop struct {
	// listener worker pool queue for polling listeners for new commands
	queue      chan Worker
	maxWorkers int
	cache      *service.Cache
	coolDown   time.Duration
}

// NewEventLoop creates a new instance of EventLoop.
func NewEventLoop(maxReceivers, maxWorkers int, coolDown time.Duration, cache *service.Cache) *EventLoop {
	queue := make(chan Worker, maxReceivers)

	return &EventLoop{queue: queue, maxWorkers: maxWorkers, coolDown: coolDown, cache: cache}
}

// AddWorker adds a new worker to the worker pool.
func (el *EventLoop) AddWorker(worker ...Worker) {
	for _, w := range worker {
		el.queue <- w
	}
}

// loop starts the event loop process.
func (el *EventLoop) loop(done chan struct{}) *sync.WaitGroup {
	workers := sync.WaitGroup{}
	for i := 0; i < el.maxWorkers; i++ {
		workers.Add(1)
		go func() {
			for {
				select {
				case <-done:
					workers.Done()
					done <- struct{}{}
					return
				case w := <-el.queue:
					el.process(w)
					go func() {
						time.Sleep(el.coolDown)
						el.queue <- w
					}()
				}
			}
		}()
	}

	return &workers
}

// process handles incoming commands.
func (el *EventLoop) process(w Worker) {
	for _, command := range *(w.Fetch()) {
		// check for "ping" requests
		if strings.EqualFold(command.Name, "ping") {
			glog.Infoln("executed \"ping\" request")
			w.Send("pong")
			continue
		}
		// ignore "pong" command relay
		if w.LoopBack() && strings.EqualFold(command.Name, "pong") {
			continue
		}

		recipient := w.Recipient()
		// check if the command request is authorized given the client number
		// that initiated it
		if !service.ClientAuthorized(command.Name, recipient) {
			glog.Warningf("%s attempted to run command \"%s\" while unauthorized to do so", recipient, command.Name)
			continue
		}

		clientPool, err := el.cache.Get(command.Name)
		if err != nil {
			glog.Warningf("invalid command \"%s\" found", command.Name)
			continue
		}
		client, ok := clientPool.Get().(*service.Service)
		if !ok {
			glog.Errorf("unable to fetch client from %s's service pool", command.Name)
			w.Send("internal error, try later")
			continue
		}
		glog.Infof("executed \"%s\" with args \"%v\"", command.Name, command.Args)
		msg, err := client.Execute(&command)
		if err != nil {
			msg = err.Error()
		}

		w.Send(msg)

		clientPool.Put(client)
	}
}

// Start begins the event loop.
func (el *EventLoop) Start(done chan struct{}, wge *sync.WaitGroup) {
	wg := el.loop(done)

	wg.Wait()
	wge.Done()
}
