package text

import (
	"sync"
	"time"
)

// Sync handles the synchronization of new texts for each of the
// listeners and pushes messages to a given callback.
type Sync struct {
	// listener worker pool queue for polling listeners for new commands
	queue      chan *Listener
	maxWorkers int
}

// workerCallback is the signature of the callback function which will
// be called when new messages appear.
type workerCallback func(tr *Listener)

// NewSync creates a new instance of Sync.
func NewSync(maxReceivers, maxWorkers int) *Sync {
	queue := make(chan *Listener, maxReceivers)

	return &Sync{queue: queue, maxWorkers: maxWorkers}
}

// AddRecipient adds a new listener to the group of command listeners.
func (s *Sync) AddRecipient(recipient ...*Listener) {
	for _, r := range recipient {
		s.queue <- r
	}
}

// Start begins the eventloop of listening for new commands given the set
// of provided listeners.
func (s *Sync) Start(wc workerCallback, done chan struct{}) *sync.WaitGroup {
	workers := sync.WaitGroup{}
	for i := 0; i < s.maxWorkers; i++ {
		workers.Add(1)
		go func() {
			for {
				select {
				case <-done:
					workers.Done()
					done <- struct{}{}
					return
				case l := <-s.queue:
					wc(l)
					go func() {
						time.Sleep(time.Duration(time.Second * 10))
						s.queue <- l
					}()
				}
			}
		}()
	}

	return &workers
}
