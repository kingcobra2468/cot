// text handles the stream of commands from client numbers and the overall
// runtime.
package text

import (
	"strings"

	"github.com/golang/glog"
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
			glog.Warningf("%s attempted to run command \"%s\" while unauthorized to do so", l.link.ClientNumber, command.Name)
			continue
		}

		clientPool, err := e.cache.Get(command.Name)
		if err != nil {
			glog.Warningf("invalid command \"%s\" found", command.Name)
			continue
		}
		client, ok := clientPool.Get().(*service.Service)
		if !ok {
			glog.Errorf("unable to fetch client from %s's service pool", command.Name)
			l.SendText("internal error, try later")
			continue
		}
		glog.Infof("executed \"%s\" with args \"%v\"", command.Name, command.Args)
		msg, err := client.Execute(&command)
		// undo any existing encoding on quotes
		msg = strings.ReplaceAll(msg, "\\\"", "\"")
		// encode all quotes
		msg = strings.ReplaceAll(msg, "\"", "\\\"")
		// remove all newlines as they cannot exist when sending messages with gvms
		msg = strings.ReplaceAll(msg, "\n", "")
		if err != nil {
			errMsg := err.Error()
			errMsg = strings.ReplaceAll(errMsg, "\\\"", "\"")
			errMsg = strings.ReplaceAll(errMsg, "\"", "\\\"")
			errMsg = strings.ReplaceAll(errMsg, "\n", "")

			l.SendText(errMsg)
		} else {
			l.SendText(msg)
		}

		clientPool.Put(client)
	}
}
