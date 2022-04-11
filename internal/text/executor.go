// text handles the stream of commands from client numbers and the overall
// runtime.
package text

import (
	"fmt"
	"strings"

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
func (e Executor) runCommand(l *Listener) {
	for _, command := range *l.Fetch() {
		// check if the command request is authorized given the client number
		// that initiated it
		if !service.ClientAuthorized(command.Name, l.link.ClientNumber) {
			fmt.Println("unauthorized request found")
			continue
		}

		clientPool, err := e.cache.Get(command.Name)
		if err != nil {
			fmt.Println("invalid command found")
			continue
		}

		client, ok := clientPool.Get().(service.Service)
		if !ok {
			continue
		}
		// TODO: make into log
		fmt.Println(command)

		message, err := client.Execute(&command)
		if err != nil {
			msg := err.Error()
			// fix encoding issues when sending errors
			msg = strings.ReplaceAll(msg, "\"", "\\\"")
			msg = strings.ReplaceAll(msg, "\n", "")
			l.SendText(msg)
		} else {
			l.SendText(message)
		}

		clientPool.Put(client)
	}
}
