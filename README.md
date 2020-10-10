# gocache

![build](https://github.com/TwinProduction/gocache/workflows/build/badge.svg?branch=master) 
[![Go Report Card](https://goreportcard.com/badge/github.com/TwinProduction/gocache)](https://goreportcard.com/report/github.com/TwinProduction/gocache)
[![Go version](https://img.shields.io/github/go-mod/go-version/TwinProduction/gocache.svg)](https://github.com/TwinProduction/gocache)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/TwinProduction/gocache)
[![Docker pulls](https://img.shields.io/docker/pulls/twinproduction/gocache-server.svg)](https://cloud.docker.com/repository/docker/twinproduction/gocache-server)

gocache is an easy-to-use, high-performance, lightweight and thread-safe (goroutine-safe) in-memory key-value cache 
with support for LRU and FIFO eviction policies as well as expiration, bulk operations and even persistence to file.


## Table of Contents

- [Features](#features)
- [Usage](#usage)
  - [Initializing the cache](#initializing-the-cache)
  - [Functions](#functions)
  - [Examples](#examples)
    - [Creating or updating an entry](#creating-or-updating-an-entry)
    - [Getting an entry](#getting-an-entry)
    - [Deleting an entry](#deleting-an-entry)
    - [Complex example](#complex-example)
  - [Persistence](#persistence)
  - [Server](#server)
- [Running the server with Docker](#running-the-server-with-docker)


## Features
gocache supports the following cache eviction policies: 
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
```go
cache := gocache.NewCache().WithMaxSize(1000).WithEvictionPolicy(gocache.LeastRecentlyUsed)
```

If you're planning on using expiration (`SetWithTTL` or `Expire`) and you want expired entries to be automatically deleted 
in the background, make sure to start the janitor when you instantiate the cache:

```go
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
| SetAll             | Same as `Set`, but in bulk
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
```go
cache.Set("key", "value") 
cache.Set("key", 1)
cache.Set("key", struct{ Text string }{Test: "value"})
```

#### Getting an entry
```go
value, ok := cache.Get("key")
```
You can also get multiple entries by using `cache.GetAll([]string{"key1", "key2"})`

#### Deleting an entry
```go
cache.Delete("key")
```
You can also delete multiple entries by using `cache.DeleteAll([]string{"key1", "key2"})`

#### Complex example
```go
package main

import (
	"fmt"
	"github.com/TwinProduction/gocache"
	"time"
)

func main() {
	cache := gocache.NewCache().WithEvictionPolicy(gocache.LeastRecentlyUsed).WithMaxSize(10000)
	cache.StartJanitor() // Manages expired entries

	cache.Set("key", "value")
	cache.SetWithTTL("key-with-ttl", "value", 60*time.Minute)
	cache.SetAll(map[string]interface{}{"k1": "v1", "k2": "v2", "k3": "v3"})

	value, exists := cache.Get("key")
	fmt.Printf("[Get] key=key; value=%s; exists=%v\n", value, exists)
	for key, value := range cache.GetAll([]string{"k1", "k2", "k3"}) {
		fmt.Printf("[GetAll] key=%s; value=%s\n", key, value)
	}
	for _, key := range cache.GetKeysByPattern("key*", 0) {
		fmt.Printf("[GetKeysByPattern] key=%s\n", key)
	}

	fmt.Println("Cache size before persisting cache to file:", cache.Count())
	err := cache.SaveToFile("cache.bak")
	if err != nil {
		panic(fmt.Sprintf("failed to persist cache to file: %s", err.Error()))
	}

	cache.Expire("key", time.Hour)
	time.Sleep(500*time.Millisecond)
	timeUntilExpiration, _ := cache.TTL("key")
	fmt.Println("Number of minutes before 'key' expires:", int(timeUntilExpiration.Seconds()))

	cache.Delete("key")
	cache.DeleteAll([]string{"k1", "k2", "k3"})

	fmt.Println("Cache size before restoring cache from file:", cache.Count())
	_, err = cache.ReadFromFile("cache.bak")
	if err != nil {
		panic(fmt.Sprintf("failed to restore cache from file: %s", err.Error()))
	}

	fmt.Println("Cache size after restoring cache from file:", cache.Count())
	cache.Clear()
	fmt.Println("Cache size after clearing the cache:", cache.Count())
}
```

<details>
  <summary>Output</summary>

```
[Get] key=key; value=value; exists=true
[GetAll] key=k2; value=v2
[GetAll] key=k3; value=v3
[GetAll] key=k1; value=v1
[GetKeysByPattern] key=key
[GetKeysByPattern] key=key-with-ttl
Cache size before persisting cache to file: 5
Number of minutes before 'key' expires: 3599
Cache size before restoring cache from file: 1
Cache size after restoring cache from file: 5
Cache size after clearing the cache: 0
```
</details>


### Persistence
While gocache is an in-memory cache, you can still save the content of the cache in a file
and vice versa.

To save the content of the cache to a file:
```go
err := cache.SaveToFile(TestCacheFile)
```

To retrieve the content of the cache from a file:
```go
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

```go
package main

import (
	"github.com/TwinProduction/gocache"
	"github.com/TwinProduction/gocache/gocacheserver"
)

func main() {
	cache := gocache.NewCache().WithEvictionPolicy(gocache.LeastRecentlyUsed).WithMaxSize(100000)
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
- [X] MGET
- [X] MSET
- [X] SCAN (kind of - cursor is not currently supported)
- [ ] KEYS


## Running the server with Docker

To build it locally, refer to the Makefile's `docker-build` and `docker-run` steps.

Note that the server version of gocache is still under development.

```
docker run --name gocache-server -p 6379:6379 twinproduction/gocache-server:v0.1.0
```
