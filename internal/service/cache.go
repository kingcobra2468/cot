package service

import "sync"

// Cache contains a goroutine-safe service client pool.
type Cache struct {
	services map[string]*sync.Pool
	mtx      sync.Mutex
}

// NewCache creates a new Cache instance.
func NewCache() *Cache {
	return &Cache{services: make(map[string]*sync.Pool), mtx: sync.Mutex{}}
}

// Add service(s) to the cache. Creates a new client pool for each service.
func (c *Cache) Add(services ...Service) {
	for _, s := range services {
		c.mtx.Lock()
		c.services[s.Name] = newServicePool(s)
		c.mtx.Unlock()
	}
}

// Services returns a list of service names.
func (c *Cache) Services() []string {
	names := make([]string, 0, len(c.services))
	for n := range c.services {
		names = append(names, n)
	}

	return names
}

// newServicePool creates a new client pool instance for a given service.
func newServicePool(s Service) *sync.Pool {
	return &sync.Pool{New: func() interface{} {
		return &Service{Name: s.Name, Domain: s.Domain}
	}}
}
