# gocache

![build](https://github.com/TwinProduction/gocache/workflows/build/badge.svg?branch=master) 
[![Go Report Card](https://goreportcard.com/badge/github.com/TwinProduction/gocache)](https://goreportcard.com/report/github.com/TwinProduction/gocache)

An extremely lightweight and minimal cache.

It supports the following cache eviction policies: 
- First in first out (FIFO)
- Least recently used (LRU)

It also supports cache entry TTL, which is both active and passive. Active expiration means that if you attempt 
to retrieve a cache key that has already expired, it will delete it on the spot and the behavior will be as if
the cache key didn't exist. As for passive expiration, there's a background task that will take care of deleting
expired keys.


## Usage
```
go get -u github.com/TwinProduction/gocache
```

### Initializing the cache
```golang
cache := gocache.NewCache().WithMaxSize(1000).WithEvictionPolicy(gocache.LeastRecentlyUsed)
```

### Creating or updating an entry
```golang
cache.Set("key", "value")
cache.Set("key", 1)
cache.Set("key", struct{ Text string }{Test: "value"})
```

### Getting an entry
```golang
value, ok := cache.Get("key")
```

You can also get multiple entries by using `cache.GetAll([]string{"key1", "key2"})`

### Deleting an entry
```
cache.Delete("key")
```

You can also delete multiple entries by using `cache.DeleteAll([]string{"key1", "key2"})`

### Persistence
While gocache is an in-memory cache, you can still save the content of the cache in a file
and vice versa.

To save the content of the cache to a file:
```golang
err := cache.SaveToFile(TestCacheFile)
```

To retrieve the content of the cache from a file:
```golang
numberOfEntriesEvicted, err := newCache.ReadFromFile(TestCacheFile)
```
The `numberOfEntriesEvicted` will be non-zero only if the number of entries 
in the file is higher than the cache's configured `MaxSize`.

### Other
```
cache.Count()
cache.Clear()
```

### Server

For the sake of convenience, a ready-to-go cache server is available 
through the `gocacheserver` package. 

The reason why the server is in a different package is because `gocache` does not use 
any external dependencies, but rather than re-inventing the wheel, the server 
implementation uses redcon, which is a Redis server framework for Go.

That way, those who desire to use gocache without the server will not add any extra dependencies
as long as they don't import the `gocacheserver` package. 

```golang
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
```

Any Redis client should be able to interact with the server, though only the following instructions are supported:
- [X] GET
- [X] SET
- [X] DEL
- [X] PING
- [X] QUIT
- [X] INFO
- [X] EXPIRE
- [X] SETEX
- [X] TTL
- [X] FLUSHDB
- [ ] KEYS
- [ ] SCAN
- [X] EXISTS
- [X] ECHO


## Running server with Docker

See the Makefile's `docker-build` and `docker-run` steps.

Note that the server version of gocache is still under development. 
