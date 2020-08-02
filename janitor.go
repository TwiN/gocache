package gocache

import (
	"log"
	"runtime"
	"time"
)

const (
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

// StartJanitor starts the janitor on a different goroutine
// The janitor's job is to delete expired keys in the background.
// It can be stopped by calling Cache.StopJanitor.
// If you do not start the janitor, expired keys will only be deleted when they are accessed through Get
func (cache *Cache) StartJanitor() error {
	if cache.stopJanitor != nil {
		return ErrJanitorAlreadyRunning
	}
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
					start := time.Now()
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
					}
					if current == cache.tail {
						if Debug {
							log.Printf("There are currently %d entries in the cache. The last walk resulted in finding %d expired keys", len(cache.entries), totalNumberOfExpiredKeysInPreviousRunFromTailToHead)
						}
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
					if Debug {
						log.Printf("traversed %d nodes and found %d expired entries in %s before stopping\n", steps, expiredEntriesFound, time.Since(start))
					}
					totalNumberOfExpiredKeysInPreviousRunFromTailToHead += expiredEntriesFound
				} else {
					if backOff*2 < JanitorMaxShiftBackOff {
						backOff *= 2
					} else {
						backOff = JanitorMaxShiftBackOff
					}
				}
				cache.mutex.Unlock()
			case <-cache.stopJanitor:
				cache.stopJanitor = nil
				return
			}
		}
	}()
	if Debug {
		go func() {
			var m runtime.MemStats
			for {
				runtime.ReadMemStats(&m)
				log.Printf("Alloc=%vMB; HeapReleased=%vMB; Sys=%vMB; HeapInUse=%vMB; HeapObjects=%v; HeapObjectsFreed=%v; GC=%v\n", m.Alloc/1024/1024, m.HeapReleased/1024/1024, m.Sys/1024/1024, m.HeapInuse/1024/1024, m.HeapObjects, m.Frees, m.NumGC)
				time.Sleep(3 * time.Second)
			}
		}()
	}
	return nil
}

// StopJanitor stops the janitor
func (cache *Cache) StopJanitor() {
	cache.stopJanitor <- true
	time.Sleep(100 * time.Millisecond)
}