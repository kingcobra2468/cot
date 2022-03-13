package text

import (
	"fmt"

	"github.com/kingcobra2468/cot/internal/service"
)

// Executor handles the synchronization of new commands for each listener.
// Also handles the propagation of fetched commands to the appropriate service
// were it will be executed.
type Executor struct {
	*Sync
	router *service.Router
}

// NewExecutor creates a new NewExecutor instance.
func NewExecutor(maxReceivers, maxWorkers int, r *service.Router) *Executor {
	sync := NewSync(maxReceivers, maxWorkers)
	return &Executor{Sync: sync, router: r}
}

// Start begins the event loop which syncs messages and executes them against
// registered services.
func (e Executor) Start(done <-chan struct{}) {
	wg := e.Sync.Start(e.runCommand, done)
	wg.Wait()
}

// runCommand fetches new commands for a given listener and then propagates
// the commands to the appropriate service input stream.
func (e Executor) runCommand(tr textRecipient) {
	for _, command := range *tr.Fetch() {
		stream, err := e.router.Get(command.Name)
		if err != nil {
			fmt.Println("found invalid command")
			continue
		}

		go func(c service.Command) {
			stream <- c
		}(command)
	}
}
