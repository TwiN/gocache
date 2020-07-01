# gocache

An extremely lightweight and minimal cache.

It supports the following cache eviction policies: 
- First in first out (FIFO)
- Least recently used (LRU)


## Usage

```
go get github.com/TwinProduction/gocache
```

```golang
cache := gocache.NewCache().WithMaxSize(1000).WithEvictionPolicy(gocache.LeastRecentlyUsed)
cache.Set("key", "value")
value, ok := cache.Get("key")
cache.Delete("key")
cache.Count()
```