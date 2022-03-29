package service

import "sync"

type Cache struct {
	services map[string]*sync.Pool
	mtx      sync.Mutex
}

func NewCache() *Cache {
	return &Cache{services: make(map[string]*sync.Pool), mtx: sync.Mutex{}}
}

func (c *Cache) Add(services ...Service) {
	for _, s := range services {
		c.mtx.Lock()
		c.services[s.Name] = newServicePool(s)
		c.mtx.Unlock()
	}
}

func (c *Cache) Services() []string {
	names := make([]string, 0, len(c.services))
	for n := range c.services {
		names = append(names, n)
	}

	return names
}

func newServicePool(s Service) *sync.Pool {
	return &sync.Pool{New: func() interface{} {
		return &Service{Name: s.Name, Domain: s.Domain}
	}}
}
