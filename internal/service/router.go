package service

import (
	"errors"
)

// Router handles the mapping between a given service and its
// input channel.
type Router struct {
	service map[string]chan Command
	Cache   *Cache
}

// NewRouter creates a new Router instance given the known
// services that were specified in the Cache.
func NewRouter(c *Cache) *Router {
	streams := make(map[string]chan Command)
	for _, name := range c.Services() {
		streams[name] = make(chan Command)
	}

	return &Router{service: streams, Cache: c}
}

// Send pushes a given Command to the correct input channel
// if such a service exists.
func (ss *Router) Send(c Command) error {
	stream, ok := ss.service[c.Name]
	if !ok {
		return errors.New("invalid command")
	}
	go func() { stream <- c }()

	return nil
}

// Get fetches the underlying channel for a given service if
// such a service exists.
func (ss *Router) Get(name string) (chan Command, error) {
	if stream, ok := ss.service[name]; ok {
		return stream, nil
	}

	return nil, errors.New("invalid service")
}
