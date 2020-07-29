package gocache

import (
	"errors"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"
)

const (
	// NoMaxSize means that the cache has no maximum number of entries in the cache
	// Setting MaxSize to this value also means there will be no eviction
	NoMaxSize = 0

	// DefaultMaxSize is the max size set if no max size is specified
	DefaultMaxSize = 1000

	// NoExpiration is the value that must be used as TTL to specify that the given key should never expire
	NoExpiration = -1

	// JanitorShiftTarget is the target number of expired keys to find during passive clean up duty
	// before pausing the passive expired keys eviction process
	JanitorShiftTarget = 25

	// JanitorMaxIterationsPerShift is the maximum number of nodes to traverse before pausing
	JanitorMaxIterationsPerShift = 1000

	// JanitorMinShiftBackOff is the minimum interval between each iteration of steps
	// defined by JanitorMaxIterationsPerShift
	JanitorMinShiftBackOff = time.Millisecond * 50

	// JanitorMaxShiftBackOff is the maximum interval between each iteration of steps
	// defined by JanitorMaxIterationsPerShift
	JanitorMaxShiftBackOff = time.Millisecond * 500
)

var (
	ErrKeyDoesNotExist       = errors.New("key does not exist")
	ErrKeyHasNoExpiration    = errors.New("key has no expiration")
	ErrJanitorAlreadyRunning = errors.New("janitor is already running")
)

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

func (cache *Cache) StartJanitor() error {
	if cache.stopJanitor != nil {
		return ErrJanitorAlreadyRunning
	}
	log.Println("[gocache] Starting Janitor")
	cache.stopJanitor = make(chan bool)
	go func() {
		// rather than starting from the tail on every run, we can try to start from the last next entry
		var lastTraversedNode *Entry
		totalNumberOfExpiredKeysInPreviousRunFromTailToHead := 0
		backOff := JanitorMinShiftBackOff
		for {
			select {
			case <-time.After(backOff):
				// Passive clean up duty
				cache.mutex.Lock()
				if cache.tail != nil {
					//start := time.Now()
					steps := 0
					expiredEntriesFound := 0
					current := cache.tail
					if lastTraversedNode != nil {
						// Make sure the lastTraversedNode is still in the cache, otherwise we might be traversing nodes that were already deleted.
						// Furthermore, we need to make sure that the entry from the cache has the same pointer as the lastTraversedNode
						// to verify that there isn't just a new cache entry with the same key (i.e. in case lastTraversedNode got evicted)
						if entryFromCache, isInCache := cache.get(lastTraversedNode.Key); isInCache && entryFromCache == lastTraversedNode {
							current = lastTraversedNode
						}
					} else {
						// XXX: DEBUG
						if cache.tail != nil && cache.tail == current {
							if _, isInCache := cache.get(cache.tail.Key); !isInCache {
								log.Println(fmt.Sprintf("starting the walk from a tail that is no longer in the cache...? key=%s; prev=%s; next=%s; tail=%s; head=%s\n", printIfNotNil(current), printIfNotNil(current.previous), printIfNotNil(current.next), printIfNotNil(cache.tail), printIfNotNil(cache.head)))
							}
						}
						// XXX: DEBUG
					}

					if current == cache.tail {
						log.Printf("There are currently %d entries in the cache. The last walk resulted in finding %d expired keys", len(cache.entries), totalNumberOfExpiredKeysInPreviousRunFromTailToHead)
						totalNumberOfExpiredKeysInPreviousRunFromTailToHead = 0
					}
					for current != nil {
						var next *Entry
						steps++
						if current.Expired() {
							expiredEntriesFound++
							// Because delete will remove the next reference from the entry, we need to store the
							// next reference before we delete it
							next = current.next
							cache.delete(current.Key)
							cache.Stats.ExpiredKeys++
						}
						if current == cache.head {
							lastTraversedNode = nil
							break
						}
						// Travel to the current node's next node only if no specific next node has been specified
						if next != nil {
							current = next
						} else {
							current = current.next
						}
						lastTraversedNode = current
						if steps == JanitorMaxIterationsPerShift || expiredEntriesFound >= JanitorShiftTarget {
							if expiredEntriesFound > 0 {
								backOff = JanitorMinShiftBackOff
							} else {
								if backOff*2 <= JanitorMaxShiftBackOff {
									backOff *= 2
								} else {
									backOff = JanitorMaxShiftBackOff
								}
							}
							break
						}
					}
					//log.Printf("traversed %d nodes and found %d expired entries in %s before stopping\n", steps, expiredEntriesFound, time.Since(start))
					totalNumberOfExpiredKeysInPreviousRunFromTailToHead += expiredEntriesFound
				} else {
					if backOff*2 < JanitorMaxShiftBackOff {
						backOff *= 2
					} else {
						backOff = JanitorMaxShiftBackOff
					}
					if len(cache.entries) > 0 {
						fmt.Println("tail is nil but cache is not empty?")
					}
				}
				cache.mutex.Unlock()
			case <-cache.stopJanitor:
				log.Println("[gocache] Stopping Janitor")
				cache.stopJanitor = nil
				return
			}
		}
	}()
	go func() {
		var m runtime.MemStats
		for {
			runtime.ReadMemStats(&m)
			fmt.Printf("Alloc=%vMB; HeapReleased=%vMB; Sys=%vMB; HeapInUse=%vMB; HeapObjects=%v; HeapObjectsFreed=%v; GC=%v\n", m.Alloc/1024/1024, m.HeapReleased/1024/1024, m.Sys/1024/1024, m.HeapInuse/1024/1024, m.HeapObjects, m.Frees, m.NumGC)
			time.Sleep(3 * time.Second)
		}
	}()
	return nil
}

func (cache *Cache) StopJanitor() {
	cache.stopJanitor <- true
	time.Sleep(100 * time.Millisecond)
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
		// so might as well just not create it in the first place    XXX: is this even necessary?
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

// Delete removes a key from the cache
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
// A ttl of -1 means that the key will never expire
// A ttl of 0 means that the key will expire immediately
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
	// if entry == cache.tail == cache.head, then won't this mean that this will just make it point to itself?
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

func printIfNotNil(entry *Entry) string {
	if entry != nil {
		return entry.Key
	}
	return "<nil>"
}
