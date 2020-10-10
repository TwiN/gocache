package gocache

import (
	"errors"
	"sync"
	"time"
)

const (
	Debug = false

	// NoMaxSize means that the cache has no maximum number of entries in the cache
	// Setting MaxSize to this value also means there will be no eviction
	NoMaxSize = 0

	// DefaultMaxSize is the max size set if no max size is specified
	DefaultMaxSize = 1000

	// NoExpiration is the value that must be used as TTL to specify that the given key should never expire
	NoExpiration = -1
)

var (
	ErrKeyDoesNotExist       = errors.New("key does not exist")
	ErrKeyHasNoExpiration    = errors.New("key has no expiration")
	ErrJanitorAlreadyRunning = errors.New("janitor is already running")
)

// Cache is the core struct of gocache which contains the data as well as all relevant configuration fields
type Cache struct {
	// MaxSize is the maximum amount of entries that can be in the cache at any given time
	MaxSize int

	// EvictionPolicy is the eviction policy
	EvictionPolicy EvictionPolicy

	Stats *Statistics

	entries map[string]*Entry
	mutex   sync.RWMutex

	head *Entry
	tail *Entry

	// stopJanitor is the channel used to stop the janitor
	stopJanitor chan bool
}

// NewCache creates a new Cache
func NewCache() *Cache {
	return &Cache{
		MaxSize:        DefaultMaxSize,
		EvictionPolicy: FirstInFirstOut,
		Stats:          &Statistics{},
		entries:        make(map[string]*Entry),
		mutex:          sync.RWMutex{},
		stopJanitor:    nil,
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
	cache.SetWithTTL(key, value, NoExpiration)
}

// SetWithTTL creates or updates a key with a given value and sets an expiration time (-1 is no expiration)
func (cache *Cache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	cache.mutex.Lock()
	entry, ok := cache.get(key)
	if !ok {
		// A negative TTL that isn't -1 (NoExpiration) is an entry that will expire instantly,
		// so might as well just not create it in the first place
		if ttl != NoExpiration && ttl < 0 {
			cache.mutex.Unlock()
			return
		}
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
		cache.entries[key] = entry
	} else {
		entry.Value = value
		// Because we just updated the entry, we need to move it back to HEAD
		cache.moveExistingEntryToHead(entry)
	}
	if ttl != NoExpiration {
		entry.Expiration = time.Now().Add(ttl).UnixNano()
	} else {
		entry.Expiration = NoExpiration
	}
	// If the cache doesn't have a MaxSize, then there's no point checking if we need to evict
	// an entry, so we'll just return now
	if cache.MaxSize == NoMaxSize {
		cache.mutex.Unlock()
		return
	}
	if len(cache.entries) > cache.MaxSize {
		cache.evict()
	}
	cache.mutex.Unlock()
}

// SetAll creates or updates multiple values
func (cache *Cache) SetAll(entries map[string]interface{}) {
	for key, value := range entries {
		cache.SetWithTTL(key, value, NoExpiration)
	}
}

// Get retrieves an entry using the key passed as parameter
// If there is no such entry, the value returned will be nil and the boolean will be false
// If there is an entry, the value returned will be the value cached and the boolean will be true
func (cache *Cache) Get(key string) (interface{}, bool) {
	cache.mutex.Lock()
	entry, ok := cache.get(key)
	if !ok {
		cache.mutex.Unlock()
		return nil, false
	}
	if entry.Expired() {
		cache.delete(key)
		cache.mutex.Unlock()
		return nil, false
	}
	if cache.EvictionPolicy == LeastRecentlyUsed {
		entry.Accessed()
		if cache.head == entry {
			cache.mutex.Unlock()
			return entry.Value, true
		}
		// Because the eviction policy is LRU, we need to move the entry back to HEAD
		cache.moveExistingEntryToHead(entry)
	}
	cache.mutex.Unlock()
	return entry.Value, true
}

// GetAll retrieves multiple entries using the keys passed as parameter
// All keys are returned in the map, regardless of whether they exist or not,
// however, entries that do not exist in the cache will return nil, meaning that
// there is no way of determining whether a key genuinely has the value nil, or
// whether it doesn't exist in the cache using only this function
func (cache *Cache) GetAll(keys []string) map[string]interface{} {
	entries := make(map[string]interface{})
	for _, key := range keys {
		entries[key], _ = cache.Get(key)
	}
	return entries
}

// GetKeysByPattern retrieves a slice of keys that match a given pattern
// i.e. cache.GetKeysByPattern("*some*") will return all keys containing "some" in them
//
// Note that GetKeysByPattern does not trigger evictions, nor does it count as accessing the entry.
func (cache *Cache) GetKeysByPattern(pattern string) []string {
	var matchingKeys []string
	cache.mutex.RLock()
	for key := range cache.entries {
		if MatchPattern(pattern, key) {
			matchingKeys = append(matchingKeys, key)
		}
	}
	cache.mutex.RUnlock()
	return matchingKeys
}

// Delete removes a key from the cache
//
// Returns false if the key did not exist.
func (cache *Cache) Delete(key string) bool {
	cache.mutex.Lock()
	ok := cache.delete(key)
	cache.mutex.Unlock()
	return ok
}

// DeleteAll deletes multiple entries based on the keys passed as parameter
//
// Returns the number of keys deleted
func (cache *Cache) DeleteAll(keys []string) int {
	numberOfKeysDeleted := 0
	cache.mutex.Lock()
	for _, key := range keys {
		if cache.delete(key) {
			numberOfKeysDeleted++
		}
	}
	cache.mutex.Unlock()
	return numberOfKeysDeleted
}

// Count returns the total amount of entries in the cache, regardless of whether they're expired or not
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

// TTL returns the time until the cache entry specified by the key passed as parameter
// will be deleted.
func (cache *Cache) TTL(key string) (time.Duration, error) {
	cache.mutex.RLock()
	entry, ok := cache.get(key)
	cache.mutex.RUnlock()
	if !ok {
		return 0, ErrKeyDoesNotExist
	}
	if entry.Expiration == NoExpiration {
		return 0, ErrKeyHasNoExpiration
	}
	timeUntilExpiration := time.Until(time.Unix(0, entry.Expiration))
	if timeUntilExpiration < 0 {
		// The key has already expired but hasn't been deleted yet.
		// From the client's perspective, this means that the cache entry doesn't exist
		return 0, ErrKeyDoesNotExist
	}
	return timeUntilExpiration, nil
}

// Expire sets a key's expiration time
//
// A TTL of -1 means that the key will never expire
// A TTL of 0 means that the key will expire immediately
// If using LRU, note that this does not reset the position of the key
//
// Returns true if the cache key exists and has had its expiration time altered
func (cache *Cache) Expire(key string, ttl time.Duration) bool {
	entry, ok := cache.get(key)
	if !ok || entry.Expired() {
		return false
	}
	if ttl != NoExpiration {
		entry.Expiration = time.Now().Add(ttl).UnixNano()
	} else {
		entry.Expiration = NoExpiration
	}
	return true
}

// get retrieves an entry using the key passed as parameter, but unlike Get, it doesn't update the access time or
// move the position of the entry to the head
func (cache *Cache) get(key string) (*Entry, bool) {
	entry, ok := cache.entries[key]
	return entry, ok
}

func (cache *Cache) delete(key string) bool {
	entry, ok := cache.entries[key]
	if ok {
		cache.removeExistingEntryReferences(entry)
		delete(cache.entries, key)
	}
	return ok
}

// moveExistingEntryToHead replaces the current cache head for an existing entry
func (cache *Cache) moveExistingEntryToHead(entry *Entry) {
	if !(entry == cache.head && entry == cache.tail) {
		cache.removeExistingEntryReferences(entry)
	}
	if entry != cache.head {
		entry.previous = cache.head
		entry.next = nil
		if cache.head != nil {
			cache.head.next = entry
		}
		cache.head = entry
	}
}

// removeExistingEntryReferences modifies the next and previous reference of an existing entry and re-links
// the next and previous entry accordingly, as well as the cache head or/and the cache tail if necessary.
// Note that it does not remove the entry from the cache, only the references.
func (cache *Cache) removeExistingEntryReferences(entry *Entry) {
	if cache.tail == entry && cache.head == entry {
		cache.tail = nil
		cache.head = nil
	} else if cache.tail == entry {
		cache.tail = cache.tail.next
	} else if cache.head == entry {
		cache.head = cache.head.previous
	}
	if entry.previous != nil {
		entry.previous.next = entry.next
	}
	if entry.next != nil {
		entry.next.previous = entry.previous
	}
	entry.next = nil
	entry.previous = nil
}

// evict removes the tail from the cache
func (cache *Cache) evict() {
	if cache.tail == nil || len(cache.entries) == 0 {
		return
	}
	if cache.tail != nil {
		oldTail := cache.tail
		cache.removeExistingEntryReferences(oldTail)
		delete(cache.entries, oldTail.Key)
		cache.Stats.EvictedKeys++
	}
}
