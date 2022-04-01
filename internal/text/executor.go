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
	cache *service.Cache
}

// NewExecutor creates a new NewExecutor instance.
func NewExecutor(maxReceivers, maxWorkers int, c *service.Cache) *Executor {
	sync := NewSync(maxReceivers, maxWorkers)
	return &Executor{Sync: sync, cache: c}
}

// Start begins the event loop which syncs messages and executes them against
// registered services.
func (e Executor) Start(done <-chan struct{}) {
	wg := e.Sync.Start(e.runCommand, done)
	wg.Wait()
}

// runCommand fetches new commands for a given listener and then propagates
// the commands to the appropriate service input stream.
func (e Executor) runCommand(tr *Listener) {
	for _, command := range *tr.Fetch() {
		if !service.ClientAuthorized(command.Name, tr.link.ClientNumber) {
			fmt.Println("unauthorized request found")
			continue
		}

		client, err := e.cache.Get(command.Name)
		if err != nil {
			fmt.Println("invalid command found")
			continue
		}

		client1, ok := client.Get().(service.Service)
		if !ok {
			continue
		}
		fmt.Println(command)
		message, err := client1.Execute(&command)
		if err != nil {
			tr.SendText(err.Error())
		} else {
			tr.SendText(message)
		}
	}
}
