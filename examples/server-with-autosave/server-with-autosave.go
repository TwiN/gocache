package main

import (
	"fmt"
	"github.com/TwinProduction/gocache"
	"github.com/TwinProduction/gocache/gocacheserver"
	"time"
)

func main() {
	cache := gocache.NewCache().WithEvictionPolicy(gocache.LeastRecentlyUsed).WithMaxSize(100000)
	cache.ReadFromFile("gocache.data")
	fmt.Println(cache.Count())
	server := gocacheserver.NewServer(cache).WithAutoSave(10*time.Second, "gocache.data")
	server.Start()
}
