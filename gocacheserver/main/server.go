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
	log.Println("PORT is set to", port)
	var maxCacheSize int
	if os.Getenv("MAX_CACHE_SIZE") == "" {
		maxCacheSize = gocache.DefaultMaxSize
	} else {
		maxCacheSize, _ = strconv.Atoi(os.Getenv("MAX_CACHE_SIZE"))
	}
	if maxCacheSize == 0 {
		log.Println("MAX_CACHE_SIZE is set to", maxCacheSize, "(no limit)")
	} else {
		log.Println("MAX_CACHE_SIZE is set to", maxCacheSize)
	}
	// TODO: support parsing data size (i.e. "5G")
	maxMemoryUsage, _ := strconv.Atoi(os.Getenv("MAX_MEMORY_USAGE"))
	if maxMemoryUsage == 0 {
		log.Println("MAX_MEMORY_USAGE is set to", maxMemoryUsage, "(disabled)")
	} else {
		log.Println("MAX_MEMORY_USAGE is set to", maxMemoryUsage)
	}
	autoSave := os.Getenv("AUTOSAVE") == "true"
	log.Println("AUTOSAVE is set to", autoSave)
	cache := gocache.NewCache().WithEvictionPolicy(gocache.LeastRecentlyUsed).WithMaxSize(maxCacheSize).WithMaxMemoryUsage(maxMemoryUsage)
	server := gocacheserver.NewServer(cache).WithPort(port)
	if autoSave {
		server = server.WithAutoSave(10*time.Minute, "/app/data/gocache.bak")
	}
	err := server.Start()
	if err != nil {
		panic(err.Error())
	}
}
