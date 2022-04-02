// service manages a single client service, the client to such service,
// as well as the whitelist on which client numbers can access it.
package service

import (
	"errors"
	"sync"
)

// Cache contains a goroutine-safe service client pool.
type Cache struct {
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

// Add service(s) to the cache. Creates a new client pool for each service.
func (c *Cache) Add(services ...Service) {
	for _, s := range services {
		c.mtx.Lock()
		c.cache[s.Name] = newServicePool(s)
		c.mtx.Unlock()
	}
}

// Get fetches the underlying service client
// such a service exists.
func (c *Cache) Get(name string) (*sync.Pool, error) {
	if client, ok := c.cache[name]; ok {
		return client, nil
	}

	return nil, errInvalidService
}

// Services returns a list of service names.
func (c *Cache) Services() []string {
	names := make([]string, 0, len(c.cache))
	for n := range c.cache {
		names = append(names, n)
	}

	return names
}

// newServicePool creates a new client pool instance for a given service.
func newServicePool(s Service) *sync.Pool {
	return &sync.Pool{New: func() interface{} {
		return Service{Name: s.Name, BaseURI: s.BaseURI}
	}}
}
