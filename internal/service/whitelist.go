package service

import (
	"github.com/patrickmn/go-cache"
)

var (
	whitelist = cache.New(cache.NoExpiration, 0)
)

func ClientAuthorized(serviceName string, clientNumber string) bool {
	_, found := whitelist.Get(serviceName + "-" + clientNumber)
	return found
}

func AddClient(serviceName string, clientNumber string) {
	whitelist.Set(serviceName+"-"+clientNumber, nil, cache.NoExpiration)
}
