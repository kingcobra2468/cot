package text

import (
	"sync"
	"time"

	"github.com/kingcobra2468/cot/internal/service"
)

// Sync handles the synchronization of new texts for each of the
// listeners and pushes messages to a given callback.
type Sync struct {
	queue      chan textRecipient
	maxWorkers int
}

// textRecipient is the signature of new data being pushed to the
// callback.
type textRecipient interface {
	Fetch() *[]service.Command
}

// workerCallback is the signature of the callback function which will
// be called when new messages appear.
type workerCallback func(tr textRecipient)

// NewSync creates a new instance of Sync.
func NewSync(maxReceivers, maxWorkers int) *Sync {
	queue := make(chan textRecipient, maxReceivers)

	return &Sync{queue: queue, maxWorkers: maxWorkers}
}

// AddRecipient adds a new listener to the group of command listeners.
func (ts *Sync) AddRecipient(tr textRecipient) {
	ts.queue <- tr
}

// Start begins the eventloop of listening for new commands given the set
// of provided listeners.
func (ts *Sync) Start(wc workerCallback, done <-chan struct{}) *sync.WaitGroup {
	workers := sync.WaitGroup{}
	for i := 0; i < ts.maxWorkers; i++ {
		workers.Add(1)
		go func() {
			for {
				select {
				case <-done:
					workers.Done()
					return
				case tr := <-ts.queue:
					wc(tr)
					go func() {
						time.Sleep(time.Duration(time.Second * 10))
						ts.queue <- tr
					}()
				}
			}
		}()
	}

	return &workers
}
