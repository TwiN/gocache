package gocache

import (
	"sync"
	"time"
)

type Cache struct {
	// MaxSize is the maximum amount of entries that can be in the cache at any given time
	MaxSize int

	// EvictionPolicy is the eviction policy
	EvictionPolicy EvictionPolicy

	entries map[string]*Entry
	mutex   sync.Mutex

	head *Entry
	tail *Entry
}

// NewCache creates a new Cache
func NewCache() *Cache {
	return &Cache{
		MaxSize:        1000,
		EvictionPolicy: FirstInFirstOut,
		entries:        make(map[string]*Entry),
		mutex:          sync.Mutex{},
	}
}

// WithMaxSize sets the maximum amount of entries that can be in the cache at any given time
func (cache *Cache) WithMaxSize(maxSize int) *Cache {
	cache.MaxSize = maxSize
	return cache
}

// WithEvictionPolicy sets eviction algorithm.
// Defaults to FirstInFirstOut (FIFO)
func (cache *Cache) WithEvictionPolicy(policy EvictionPolicy) *Cache {
	cache.EvictionPolicy = policy
	return cache
}

// Set creates or updates a key with a given value
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
	if !(entry == cache.head && entry == cache.tail) {
		cache.removeExistingEntry(entry)
	}
	if entry != cache.head {
		entry.previous = cache.head
		entry.next = nil
		cache.head.next = entry
		cache.head = entry
	}
}

func (cache *Cache) removeExistingEntry(entry *Entry) {
	if cache.tail == entry {
		cache.tail = cache.tail.next
	}
	if cache.head == entry {
		cache.head = entry.previous
	}
	if entry.previous != nil {
		entry.previous.next = entry.next
	}
	if entry.next != nil {
		entry.next.previous = entry.previous
	}
}

func (cache *Cache) evict() {
	cache.mutex.Lock()
	if cache.tail == nil || len(cache.entries) == 0 {
		cache.mutex.Unlock()
		return
	}
	if cache.tail != nil {
		delete(cache.entries, cache.tail.Key)
		cache.tail = cache.tail.next
		cache.tail.previous = nil
	}
	cache.mutex.Unlock()
}

func (cache *Cache) Delete(key string) {
	cache.mutex.Lock()
	entry, ok := cache.entries[key]
	if ok {
		cache.removeExistingEntry(entry)
		delete(cache.entries, key)
	}
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

// Count returns the total amount of entries in the cache
func (cache *Cache) Count() int {
	cache.mutex.Lock()
	count := len(cache.entries)
	cache.mutex.Unlock()
	return count
}

// Clear deletes all entries from the cache
func (cache *Cache) Clear() {
	cache.mutex.Lock()
	cache.entries = make(map[string]*Entry)
	cache.head = nil
	cache.tail = nil
	cache.mutex.Unlock()
}
