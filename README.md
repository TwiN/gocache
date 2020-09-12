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

While meant to be used as a library, there's a Redis-compatible cache server included. 
See the [Server](#server) section. 
It may also serve as a good reference to use in order to implement gocache in your own applications.


## Usage
```
go get -u github.com/TwinProduction/gocache
```

### Initializing the cache
```golang
cache := gocache.NewCache().WithMaxSize(1000).WithEvictionPolicy(gocache.LeastRecentlyUsed)
```

If you're planning on using expiration (`SetWithTTL` or `Expire`) and you want expired entries to be automatically deleted 
in the background, make sure to start the janitor when you instantiate the cache:

```golang
cache.StartJanitor()
```

### Functions

| Function           | Description |
| ------------------ | ----------- |
| WithMaxSize        | Sets the max size of the cache. `gocache.NoMaxSize` means there is no limit. If not set, the default max size is `gocache.DefaultMaxSize`.
| WithEvictionPolicy | Sets the eviction algorithm to be used when the cache reaches the max size. If not set, the default eviction policy is `gocache.FirstInFirstOut` (FIFO).
| StartJanitor       | Starts the janitor, which is in charge of deleting expired cache entries in the background.
| StopJanitor        | Stops the janitor.
| Set                | Same as `SetWithTTL`, but with no expiration (`gocache.NoExpiration`)
| SetWithTTL         | Creates or updates a cache entry with the given key, value and expiration time. If the max size after the aforementioned operation is above the configured max size, the tail will be evicted. Depending on the eviction policy, the tail is defined as the oldest 
| Get                | Gets a cache entry by its key.
| GetAll             | Gets a map of entries by their keys. The resulting map will contain all keys, even if some of the keys in the slice passed as parameter were not present in the cache.  
| GetKeysByPattern   | Retrieves a slice of keys that matches a given pattern.
| Delete             | Removes a key from the cache.
| DeleteAll          | Removes multiple keys from the cache.
| Count              | Gets the size of the cache. This includes cache keys which may have already expired, but have not been removed yet.
| Clear              | Wipes the cache.
| TTL                | Gets the time until a cache key expires. 
| Expire             | Sets the expiration time of an existing cache key.
| SaveToFile         | Stores the content of the cache to a file so that it can be read using `ReadFromFile`
| ReadFromFile       | Populates the cache using a file created using `SaveToFile`

### Examples

#### Creating or updating an entry
```golang
cache.Set("key", "value") 
cache.Set("key", 1)
cache.Set("key", struct{ Text string }{Test: "value"})
```

#### Getting an entry
```golang
value, ok := cache.Get("key")
```

You can also get multiple entries by using `cache.GetAll([]string{"key1", "key2"})`

#### Deleting an entry
```golang
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
- [X] EXISTS
- [X] ECHO
- [ ] KEYS
- [ ] SCAN


## Running the server with Docker

See the Makefile's `docker-build` and `docker-run` steps.

Note that the server version of gocache is still under development. 
