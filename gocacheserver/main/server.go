package main

import (
	"github.com/TwinProduction/gocache"
	"github.com/TwinProduction/gocache/gocacheserver"
	"log"
	"os"
	"strconv"
	"time"
)

func main() {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	if port == 0 {
		port = gocacheserver.DefaultServerPort
	}
	maxCacheSize, _ := strconv.Atoi(os.Getenv("MAX_CACHE_SIZE"))
	if maxCacheSize == 0 {
		maxCacheSize = gocache.DefaultMaxSize
	}
	autoSave := os.Getenv("AUTOSAVE") == "true"
	log.Println("PORT is set to", port)
	log.Println("MAX_CACHE_SIZE is set to", maxCacheSize)
	log.Println("AUTOSAVE is set to", autoSave)
	cache := gocache.NewCache().WithEvictionPolicy(gocache.LeastRecentlyUsed).WithMaxSize(maxCacheSize)
	server := gocacheserver.NewServer(cache).WithPort(port)
	if autoSave {
		server = server.WithAutoSave(10*time.Minute, "/app/data/gocache.bak")
	}
	err := server.Start()
	if err != nil {
		panic(err.Error())
	}
}
