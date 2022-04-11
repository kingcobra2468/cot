package service

import (
	"github.com/patrickmn/go-cache"
)

var (
	// registers a client number as well as the "{serviceName}-{clientNumber}"
	// hash for constant-time lookup
	whitelist = cache.New(cache.NoExpiration, 0)
)

// ClientAuthorized checks if a client number is authorized to use a given
// service.
func ClientAuthorized(serviceName string, clientNumber string) bool {
	_, found := whitelist.Get(serviceName + "-" + clientNumber)
	return found
}

// ClientExists checks if a client number has previously been recorded.
func ClientExists(clientNumber string) bool {
	_, found := whitelist.Get(clientNumber)
	return found
}

// AddClient adds a client number to a given service's whitelist. Also registers
// a client number as a known number.
func AddClient(serviceName string, clientNumber string) {
	// links a client number with a service
	whitelist.Set(serviceName+"-"+clientNumber, nil, cache.NoExpiration)
	// registers a client number as known
	whitelist.Set(clientNumber, nil, cache.NoExpiration)
}
