package gocache

import (
	"sync"
	"time"
)

type Cache struct {
	MaxSize        int
	EvictionPolicy EvictionPolicy

	entries map[string]*Entry
	mutex   sync.Mutex
}

func NewCache() *Cache {
	return &Cache{
		MaxSize:        1000,
		EvictionPolicy: FirstInFirstOut,
		entries:        make(map[string]*Entry),
		mutex:          sync.Mutex{},
	}
}

func (cache *Cache) WithMaxSize(maxSize int) *Cache {
	cache.MaxSize = maxSize
	return cache
}

func (cache *Cache) WithEvictionPolicy(policy EvictionPolicy) *Cache {
	cache.EvictionPolicy = policy
	return cache
}

func (cache *Cache) Set(key string, value interface{}) {
	cache.mutex.Lock()
	cache.entries[key] = &Entry{
		Value:             value,
		RelevantTimestamp: time.Now(),
	}
	cacheSize := len(cache.entries)
	cache.mutex.Unlock()
	if cacheSize > cache.MaxSize {
		cache.evict()
	}
}

func (cache *Cache) evict() {
	cache.mutex.Lock()
	if len(cache.entries) == 0 {
		return
	}
	var oldestKey string
	oldestKeyTimestamp := time.Now()
	for k, v := range cache.entries {
		if len(oldestKey) == 0 || oldestKeyTimestamp.After(v.RelevantTimestamp) {
			oldestKey = k
			oldestKeyTimestamp = v.RelevantTimestamp
		}
	}
	delete(cache.entries, oldestKey)
	cache.mutex.Unlock()
}

func (cache *Cache) Delete(key string) {
	cache.mutex.Lock()
	delete(cache.entries, key)
	cache.mutex.Unlock()
}

func (cache *Cache) Get(key string) (interface{}, bool) {
	cache.mutex.Lock()
	entry, ok := cache.entries[key]
	if ok && cache.EvictionPolicy == LeastRecentlyUsed {
		cache.entries[key].Accessed()
	}
	cache.mutex.Unlock()
	if !ok {
		return nil, false
	}
	return entry.Value, true
}

func (cache *Cache) Count() int {
	cache.mutex.Lock()
	count := len(cache.entries)
	cache.mutex.Unlock()
	return count
}

func (cache *Cache) Clear() {
	cache.mutex.Lock()
	cache.entries = make(map[string]*Entry)
	cache.mutex.Unlock()
}
