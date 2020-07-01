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

	head *Entry
	tail *Entry
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
	entry, ok := cache.entries[key]
	if !ok {
		// Cache entry doesn't exist, so we have to create a new one
		entry = &Entry{
			Key:               key,
			Value:             value,
			RelevantTimestamp: time.Now(),
			previous:          cache.head,
		}
		if cache.head == nil {
			cache.tail = entry
		} else {
			cache.head.next = entry
		}
		cache.head = entry
	} else {
		entry.Value = value
		// Because we just updated the entry, we need to move it back to HEAD
		cache.moveExistingEntryToHead(entry)
	}
	cache.entries[key] = entry
	cacheSize := len(cache.entries)
	cache.mutex.Unlock()
	if cacheSize > cache.MaxSize {
		cache.evict()
	}
}

func (cache *Cache) moveExistingEntryToHead(entry *Entry) {
	if cache.tail == entry {
		cache.tail = cache.tail.next
	}
	if entry.previous != nil {
		entry.previous.next = entry.next
	}
	if entry.next != nil {
		entry.next.previous = entry.previous
	}
	entry.next = nil
	entry.previous = cache.head
	cache.head.next = entry
	cache.head = entry
}

func (cache *Cache) evict() {
	cache.mutex.Lock()
	if len(cache.entries) == 0 {
		return
	}
	//var oldestKey string
	//oldestKeyTimestamp := time.Now()
	//for k, v := range cache.entries {
	//	if len(oldestKey) == 0 || oldestKeyTimestamp.After(v.RelevantTimestamp) {
	//		oldestKey = k
	//		oldestKeyTimestamp = v.RelevantTimestamp
	//	}
	//}
	//delete(cache.entries, oldestKey)

	if cache.tail != nil {
		delete(cache.entries, cache.tail.Key)
		cache.tail = cache.tail.next
		cache.tail.previous = nil
	}

	cache.mutex.Unlock()
}

func (cache *Cache) Delete(key string) {
	cache.mutex.Lock()
	delete(cache.entries, key)
	cache.mutex.Unlock()
}

// Get retrieves an entry using the key passed as parameter
// If there is no such entry, the value returned will be nil and the boolean will be false
// If there is an entry, the value returned will be the value cached and the boolean will be true
func (cache *Cache) Get(key string) (interface{}, bool) {
	cache.mutex.Lock()
	entry, ok := cache.entries[key]
	cache.mutex.Unlock()
	if !ok {
		return nil, false
	}
	if cache.EvictionPolicy == LeastRecentlyUsed {
		entry.Accessed()
		if cache.head == entry {
			return entry.Value, true
		}
		// Because the eviction policy is LRU, we need to move the entry back to HEAD
		cache.moveExistingEntryToHead(entry)
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
