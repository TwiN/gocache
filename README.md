# gocache

![build](https://github.com/TwinProduction/gocache/workflows/build/badge.svg?branch=master) 
[![Go Report Card](https://goreportcard.com/badge/github.com/TwinProduction/gocache)](https://goreportcard.com/report/github.com/TwinProduction/gocache)

An extremely lightweight and minimal cache.

It supports the following cache eviction policies: 
- First in first out (FIFO)
- Least recently used (LRU)


## Usage
```
go get -u github.com/TwinProduction/gocache
```

### Initializing the cache
```
cache := gocache.NewCache().WithMaxSize(1000).WithEvictionPolicy(gocache.LeastRecentlyUsed)
```

### Creating or updating an entry
```
cache.Set("key", "value")
```

### Getting an entry
```
value, ok := cache.Get("key")
```

### Deleting an entry
```
cache.Delete("key")
```

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
- [ ] INFO
- [ ] KEYS
- [X] EXISTS
- [X] ECHO


## Performance

Using the following command:
```
redis-benchmark -p 6379 -t set,get -n 10000000 -q -P 512 -c 512
```

On a machine with the following specs:
```
Arch Linux
x86_64 Linux 5.7.7-arch1-1
i7-8550U 8x 4GHz
16G RAM
```

### Gocache

#### Without eviction

With the following configuration:
```
cache := gocache.NewCache().WithEvictionPolicy(gocache.LeastRecentlyUsed).WithMaxSize(10000000)
server := gocacheserver.NewServer(cache)
server.Start()
```

Single-threaded (`GOMAXPROCS=1 go run examples/server.go`):
```
SET: 2239727.50 requests per second
GET: 2681068.00 requests per second
```

Multi-threaded (`go run examples/server.go`):
```
SET: 2573476.00 requests per second
GET: 6399190.50 requests per second
```

#### With eviction

With the following configuration:
```
cache := gocache.NewCache().WithEvictionPolicy(gocache.LeastRecentlyUsed).WithMaxSize(10000)
server := gocacheserver.NewServer(cache)
server.Start()
```

Single-threaded (`GOMAXPROCS=1 go run examples/server.go`):
```
SET: 2305298.50 requests per second
GET: 2745096.00 requests per second
```

Multi-threaded (`go run examples/server.go`):
```
SET: 2576740.00 requests per second
GET: 6451397.00 requests per second
```


### Redis

Using the default configuration with `redis-server`:
```
SET: 2105156.50 requests per second
GET: 2842900.75 requests per second
```

