package main

import (
	"github.com/TwinProduction/gocache"
	"github.com/TwinProduction/gocache/gocacheserver"
)

func main() {
	cache := gocache.NewCache()
	server := gocacheserver.NewServer(cache)
	server.Start()
}
