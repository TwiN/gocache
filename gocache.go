package gocache

import (
	"bufio"
	"encoding/gob"
	"os"
	"sort"
	"sync"
	"time"
)

const (
	// NoMaxSize means that the cache has no maximum number of entries in the cache
	// Setting MaxSize to this value also means there will be no eviction
	NoMaxSize = 0

	// DefaultMaxSize is the max size set if no max size is specified
	DefaultMaxSize = 1000
)

type Cache struct {
	// MaxSize is the maximum amount of entries that can be in the cache at any given time
	MaxSize int

	// EvictionPolicy is the eviction policy
	EvictionPolicy EvictionPolicy

	entries map[string]*Entry
	mutex   sync.RWMutex

	head *Entry
	tail *Entry
}

// NewCache creates a new Cachealso
func NewCache() *Cache {
	return &Cache{
		MaxSize:        DefaultMaxSize,
		EvictionPolicy: FirstInFirstOut,
		entries:        make(map[string]*Entry),
		mutex:          sync.RWMutex{},
	}
}

// WithMaxSize sets the maximum amount of entries that can be in the cache at any given time
// A MaxSize of 0 or less means infinite
func (cache *Cache) WithMaxSize(maxSize int) *Cache {
	if maxSize < 0 {
		maxSize = NoMaxSize
	}
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
	// If the cache doesn't have a MaxSize, then there's no point checking if we need to evict
	// an entry, so we'll just return now
	if cache.MaxSize == NoMaxSize {
		cache.mutex.Unlock()
		return
	}
	cacheSize := len(cache.entries)
	if cacheSize > cache.MaxSize {
		cache.evict()
	}
	cache.mutex.Unlock()
}

// Get retrieves an entry using the key passed as parameter
// If there is no such entry, the value returned will be nil and the boolean will be false
// If there is an entry, the value returned will be the value cached and the boolean will be true
func (cache *Cache) Get(key string) (interface{}, bool) {
	cache.mutex.RLock()
	entry, ok := cache.entries[key]
	cache.mutex.RUnlock()
	if !ok {
		return nil, false
	}
	if cache.EvictionPolicy == LeastRecentlyUsed {
		entry.Accessed()
		if cache.head == entry {
			return entry.Value, true
		}
		// Because the eviction policy is LRU, we need to move the entry back to HEAD
		// XXX: the following lock really hurts perf. Perhaps we should create a mutex specifically for head/tail?
		cache.mutex.Lock()
		cache.moveExistingEntryToHead(entry)
		cache.mutex.Unlock()
	}
	return entry.Value, true
}

// Delete removes a key from the cache
// Returns false if the key did not exist.
func (cache *Cache) Delete(key string) bool {
	cache.mutex.Lock()
	entry, ok := cache.entries[key]
	if ok {
		cache.removeExistingEntry(entry)
		delete(cache.entries, key)
	}
	cache.mutex.Unlock()
	return ok
}

// Count returns the total amount of entries in the cache
func (cache *Cache) Count() int {
	cache.mutex.RLock()
	count := len(cache.entries)
	cache.mutex.RUnlock()
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

// SaveToFile stores the content of the cache to a file so that it can be read using
// the ReadFromFile function
func (cache *Cache) SaveToFile(path string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	encoder := gob.NewEncoder(writer)
	cache.mutex.RLock()
	err = encoder.Encode(cache.entries)
	cache.mutex.RUnlock()
	if err != nil {
		return err
	}
	return writer.Flush()
}

// ReadFromFile populates the cache using a file created using cache.SaveToFile(path)
//
// Note that if the number of entries retrieved from the file exceed the configured MaxSize,
// the extra entries will be automatically evicted according to the EvictionPolicy configured.
// This function returns the number of entries evicted, and because this function only reads
// from a file and does not modify it, you can safely retry this function after configuring
// the cache with the appropriate MaxSize, should you desire to.
func (cache *Cache) ReadFromFile(path string) (int, error) {
	file, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	decoder := gob.NewDecoder(reader)
	cache.mutex.Lock()
	err = decoder.Decode(&cache.entries)
	if err != nil {
		return 0, err
	}
	// Because pointers don't get stored in the file, we need to relink everything from head to tail
	var entries []*Entry
	for _, v := range cache.entries {
		entries = append(entries, v)
	}
	// Sort the slice of entries from oldest to newest
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].RelevantTimestamp.Before(entries[j].RelevantTimestamp)
	})
	// Relink the nodes from tail to head
	var previous *Entry
	for i := range entries {
		current := entries[i]
		if previous == nil {
			cache.tail = current
			cache.head = current
		} else {
			previous.next = current
			current.previous = previous
			cache.head = current
		}
		previous = entries[i]
	}
	// If the cache doesn't have a MaxSize, then there's no point checking if we need to evict
	// an entry, so we'll just return now
	if cache.MaxSize == NoMaxSize {
		cache.mutex.Unlock()
		return 0, nil
	}
	// Evict until the total number of entries matches the cache's maximum size
	numberOfEvictions := 0
	for len(cache.entries) > cache.MaxSize {
		numberOfEvictions++
		cache.evict()
	}
	cache.mutex.Unlock()
	return numberOfEvictions, nil
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
	if cache.tail == nil || len(cache.entries) == 0 {
		return
	}
	if cache.tail != nil {
		delete(cache.entries, cache.tail.Key)
		cache.tail = cache.tail.next
		cache.tail.previous = nil
	}
}
