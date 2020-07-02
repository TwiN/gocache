# gocache

![build](https://github.com/TwinProduction/gocache/workflows/build/badge.svg?branch=master) 
[![Go Report Card](https://goreportcard.com/badge/github.com/TwinProduction/gocache)](https://goreportcard.com/report/github.com/TwinProduction/gocache)

An extremely lightweight and minimal cache.

It supports the following cache eviction policies: 
- First in first out (FIFO)
- Least recently used (LRU)


## Usage
```
go get github.com/TwinProduction/gocache
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

### Other
```
cache.Count()
cache.Clear()
```