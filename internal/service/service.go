package service

// Service pushes a given command to its intended service.
type Service struct {
	Name   string
	Domain string
}

// Command sets up the schema for a given command.
type Command struct {
	Name      string
	Arguments []string
}

// Listen attends to a given channel for new commands and executes
// them against a given service.
func (s Service) Listen(stream <-chan *Command, done <-chan struct{}) {
	go func() {
		for {
			select {
			case <-done:
				return
			case command := <-stream:
				s.execute(command)
			}
		}
	}()
}

// execute will run the command against the service.
func (s Service) execute(c *Command) error {
	return nil
}
