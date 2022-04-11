// service manages a single client service, the client to such service,
// as well as the whitelist on which client numbers can access it.
package service

import (
	"errors"
	"sync"
)

// Cache contains a goroutine-safe client pool for services. This is
// then used when sending commands to the client services efficiently
// as a minimum amount of clients are created via the use of a pool.
type Cache struct {
	// store a pool that can be referenced by the service/command name
	cache map[string]*sync.Pool
	mtx   sync.Mutex
}

var (
	errInvalidService = errors.New("invalid service")
)

// NewCache creates a new Cache instance.
func NewCache() *Cache {
	return &Cache{cache: make(map[string]*sync.Pool), mtx: sync.Mutex{}}
}

// Add service(s) to the cache. This creates a new client pool for each service.
// This method is goroutine-safe.
func (c *Cache) Add(services ...Service) {
	for _, s := range services {
		c.mtx.Lock()
		c.cache[s.Name] = newServicePool(s)
		c.mtx.Unlock()
	}
}

// Get fetches the underlying service under the name provided in configuration if
// such a service exists.
func (c *Cache) Get(name string) (*sync.Pool, error) {
	if client, ok := c.cache[name]; ok {
		return client, nil
	}

	return nil, errInvalidService
}

// Services fetches a list of service names registed within the Cache.
func (c *Cache) Services() []string {
	names := make([]string, 0, len(c.cache))
	for name := range c.cache {
		names = append(names, name)
	}

	return names
}

// newServicePool creates a new client pool for a given service.
func newServicePool(s Service) *sync.Pool {
	return &sync.Pool{New: func() interface{} {
		return Service{Name: s.Name, BaseURI: s.BaseURI}
	}}
}
