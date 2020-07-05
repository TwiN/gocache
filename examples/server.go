package main

import (
	"github.com/TwinProduction/gocache"
	"github.com/TwinProduction/gocache/gocacheserver"
)

func main() {
	cache := gocache.NewCache().WithEvictionPolicy(gocache.LeastRecentlyUsed).WithMaxSize(10000)
	server := gocacheserver.NewServer(cache)
	server.Start()
}
