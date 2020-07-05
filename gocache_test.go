package gocache

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"
)

const (
	TestCacheFile = "test.cache"
)

func TestCache_Get(t *testing.T) {
	cache := NewCache().WithMaxSize(10)
	cache.Set("key", "value")
	value, ok := cache.Get("key")
	if !ok {
		t.Error("expected key to exist")
	}
	if value != "value" {
		t.Errorf("expected: %s, but got: %s", "value", value)
	}
}

func TestCache_Set(t *testing.T) {
	cache := NewCache().WithMaxSize(10)
	cache.Set("key", "value")
	value, ok := cache.Get("key")
	if !ok {
		t.Error("expected key to exist")
	}
	if value != "value" {
		t.Errorf("expected: %s, but got: %s", "value", value)
	}
	cache.Set("key", "newvalue")
	value, ok = cache.Get("key")
	if !ok {
		t.Error("expected key to exist")
	}
	if value != "newvalue" {
		t.Errorf("expected: %s, but got: %s", "newvalue", value)
	}
}

func TestCache_EvictionsRespectMaxSize(t *testing.T) {
	cache := NewCache().WithMaxSize(5)
	for n := 0; n < 10; n++ {
		cache.Set(fmt.Sprintf("test_%d", n), []byte("value"))
	}
	count := cache.Count()
	if count > 5 {
		t.Error("Max size was set to 5, but the cache size reached a size of", count)
	}
}

func TestCache_EvictionsWithFIFO(t *testing.T) {
	cache := NewCache().WithMaxSize(3).WithEvictionPolicy(FirstInFirstOut)

	cache.Set("1", []byte("value"))
	time.Sleep(time.Millisecond)
	cache.Set("2", []byte("value"))
	time.Sleep(time.Millisecond)
	cache.Set("3", []byte("value"))
	time.Sleep(time.Millisecond)
	_, _ = cache.Get("1")
	cache.Set("4", []byte("value"))
	_, ok := cache.Get("1")
	if ok {
		t.Error("expected key 1 to have been removed, because FIFO")
	}
}

func TestCache_EvictionsWithLRU(t *testing.T) {
	cache := NewCache().WithMaxSize(3).WithEvictionPolicy(LeastRecentlyUsed)

	cache.Set("1", []byte("value"))
	time.Sleep(time.Millisecond)
	cache.Set("2", []byte("value"))
	time.Sleep(time.Millisecond)
	cache.Set("3", []byte("value"))
	time.Sleep(time.Millisecond)
	_, _ = cache.Get("1")
	cache.Set("4", []byte("value"))

	_, ok := cache.Get("1")
	if !ok {
		t.Error("expected key 1 to still exist, because LRU")
	}
}

func TestCache_HeadTailWorksWithFIFO(t *testing.T) {
	cache := NewCache().WithMaxSize(3).WithEvictionPolicy(FirstInFirstOut)

	if cache.tail != nil {
		t.Error("cache tail should have been nil")
	}
	if cache.head != nil {
		t.Error("cache head should have been nil")
	}

	cache.Set("1", []byte("value"))

	// (tail) 1 (head)
	if cache.tail == nil || cache.tail.Key != "1" {
		t.Error("cache tail should have been entry with key 1")
	}
	if cache.head == nil || cache.head.Key != "1" {
		t.Error("cache head should have been entry with key 1")
	}

	time.Sleep(time.Millisecond)
	cache.Set("2", []byte("value"))

	// (tail) 1 - 2 (head)
	if cache.tail == nil || cache.tail.Key != "1" {
		t.Error("cache tail should have been the entry with key 1")
	}
	if cache.head == nil || cache.head.Key != "2" {
		t.Error("cache head should have been the entry with key 2")
	}
	if cache.head.previous.Key != "1" {
		t.Error("The entry key before the cache head should have been 1")
	}
	if cache.head.next != nil {
		t.Error("The cache head should not have a next node")
	}
	if cache.tail.next.Key != "2" {
		t.Error("The entry key after the cache tail should have been 2")
	}
	if cache.tail.previous != nil {
		t.Error("The cache tail should not have a previous node")
	}

	time.Sleep(time.Millisecond)
	cache.Set("3", []byte("value"))

	// (tail) 1 - 2 - 3 (head)
	if cache.tail == nil || cache.tail.Key != "1" {
		t.Error("cache tail should have been the entry with key 1")
	}
	if cache.tail.next.Key != "2" {
		t.Error("The entry key after the cache tail should have been 2")
	}
	if cache.tail.previous != nil {
		t.Error("The cache tail should not have a previous node")
	}
	if cache.head == nil || cache.head.Key != "3" {
		t.Error("cache head should have been the entry with key 3")
	}
	if cache.head.previous.Key != "2" {
		t.Error("The entry key before the cache head should have been 2")
	}
	if cache.head.next != nil {
		t.Error("The cache head should not have a next node")
	}
	if cache.head.previous.next.Key != "3" {
		t.Error("The head's previous node should have its next node pointing to the cache head")
	}
	if cache.head.previous.previous.Key != "1" {
		t.Error("The head's previous node should have its previous node pointing to the cache tail")
	}

	time.Sleep(time.Millisecond)

	// Get the first entry. This doesn't change anything for FIFO, but for LRU, it would mean that retrieved entry
	// wouldn't be evicted since it was recently accessed. Basically, we just want to make sure that FIFO works
	// as intended (i.e. not like LRU)
	_, _ = cache.Get("1")

	cache.Set("4", []byte("value"))

	// (tail) 2 - 3 - 4 (head)
	_, ok := cache.Get("1")
	if ok {
		t.Error("expected key 1 to have been removed, because FIFO")
	}
	if cache.tail == nil || cache.tail.Key != "2" {
		t.Error("cache tail should have been the entry with key 2")
	}
	if cache.tail.next.Key != "3" {
		t.Error("The entry key after the cache tail should have been 3")
	}
	if cache.tail.previous != nil {
		t.Error("The cache tail should not have a previous node")
	}
	if cache.head == nil || cache.head.Key != "4" {
		t.Error("cache head should have been the entry with key 4")
	}
	if cache.head.previous.Key != "3" {
		t.Error("The entry key before the cache head should have been 3")
	}
	if cache.head.next != nil {
		t.Error("The cache head should not have a next node")
	}
	if cache.head.previous.next.Key != "4" {
		t.Error("The head's previous node should have its next node pointing to the cache head")
	}
	if cache.head.previous.previous.Key != "2" {
		t.Error("The head's previous node should have its previous node pointing to the cache tail")
	}
}

func TestCache_HeadTailWorksWithLRU(t *testing.T) {
	cache := NewCache().WithMaxSize(3).WithEvictionPolicy(LeastRecentlyUsed)

	if cache.tail != nil {
		t.Error("cache tail should have been nil")
	}
	if cache.head != nil {
		t.Error("cache head should have been nil")
	}

	cache.Set("1", []byte("value"))

	// (tail) 1 (head)
	if cache.tail == nil || cache.tail.Key != "1" {
		t.Error("cache tail should have been entry with key 1")
	}
	if cache.head == nil || cache.head.Key != "1" {
		t.Error("cache head should have been entry with key 1")
	}

	time.Sleep(time.Millisecond)
	cache.Set("2", []byte("value"))

	// (tail) 1 - 2 (head)
	if cache.tail == nil || cache.tail.Key != "1" {
		t.Error("cache tail should have been the entry with key 1")
	}
	if cache.head == nil || cache.head.Key != "2" {
		t.Error("cache head should have been the entry with key 2")
	}
	if cache.head.previous.Key != "1" {
		t.Error("The entry key before the cache head should have been 1")
	}
	if cache.head.next != nil {
		t.Error("The cache head should not have a next node")
	}
	if cache.tail.next.Key != "2" {
		t.Error("The entry key after the cache tail should have been 2")
	}
	if cache.tail.previous != nil {
		t.Error("The cache tail should not have a previous node")
	}

	time.Sleep(time.Millisecond)
	cache.Set("3", []byte("value"))

	// (tail) 1 - 2 - 3 (head)
	if cache.tail == nil || cache.tail.Key != "1" {
		t.Error("cache tail should have been the entry with key 1")
	}
	if cache.tail.next.Key != "2" {
		t.Error("The entry key after the cache tail should have been 2")
	}
	if cache.tail.previous != nil {
		t.Error("The cache tail should not have a previous node")
	}
	if cache.head == nil || cache.head.Key != "3" {
		t.Error("cache head should have been the entry with key 3")
	}
	if cache.head.previous.Key != "2" {
		t.Error("The entry key before the cache head should have been 2")
	}
	if cache.head.next != nil {
		t.Error("The cache head should not have a next node")
	}
	if cache.head.previous.next.Key != "3" {
		t.Error("The head's previous node should have its next node pointing to the cache head")
	}
	if cache.head.previous.previous.Key != "1" {
		t.Error("The head's previous node should have its previous node pointing to the cache tail")
	}

	time.Sleep(time.Millisecond)

	// Because we're using a LRU cache, this should cause 1 to get moved back to the head, thus
	// moving it from the tail.
	// In other words, because we retrieved the key 1 here, this is no longer the least recently used cache entry,
	// which means it will not be evicted during the next insertion.
	_, _ = cache.Get("1")

	// (tail) 2 - 3 - 1 (head) (This updated because LRU)
	cache.Set("4", []byte("value"))

	// (tail) 3 - 1 - 4 (head)
	if cache.tail == nil || cache.tail.Key != "3" {
		t.Error("cache tail should have been the entry with key 3")
	}
	if cache.tail.next.Key != "1" {
		t.Error("The entry key after the cache tail should have been 1")
	}
	if cache.tail.previous != nil {
		t.Error("The cache tail should not have a previous node")
	}
	if cache.head == nil || cache.head.Key != "4" {
		t.Error("cache head should have been the entry with key 4")
	}
	if cache.head.previous.Key != "1" {
		t.Error("The entry key before the cache head should have been 1")
	}
	if cache.head.next != nil {
		t.Error("The cache head should not have a next node")
	}
	if cache.head.previous.next.Key != cache.head.Key {
		t.Error("The head's previous node should have its next node pointing to the cache head")
	}
	if cache.head.previous.previous.Key != cache.tail.Key {
		t.Error("Should be able to walk from head to tail")
	}
	if cache.tail.next.next != cache.head {
		t.Error("Should be able to walk from tail to head")
	}

	_, ok := cache.Get("1")
	if !ok {
		t.Error("expected key 1 to still exist, because LRU")
	}
}

func TestCache_Delete(t *testing.T) {
	cache := NewCache()

	if cache.tail != nil {
		t.Error("cache tail should have been nil")
	}
	if cache.head != nil {
		t.Error("cache head should have been nil")
	}

	cache.Set("1", []byte("1"))
	time.Sleep(time.Millisecond)
	cache.Set("2", []byte("2"))
	time.Sleep(time.Millisecond)
	cache.Set("3", []byte("3"))

	// (tail) 1 - 2 - 3 (head)
	if cache.tail.Key != "1" {
		t.Error("cache tail should have been the entry with key 1")
	}
	if cache.head.Key != "3" {
		t.Error("cache head should have been the entry with key 3")
	}

	cache.Delete("2")

	// (tail) 1 - 3 (head)
	if cache.tail.Key != "1" {
		t.Error("cache tail should have been the entry with key 1")
	}
	if cache.head.Key != "3" {
		t.Error("cache head should have been the entry with key 3")
	}
	if cache.tail.next.Key != "3" {
		t.Error("The entry key after the cache tail should have been 3")
	}
	if cache.head.previous.Key != "1" {
		t.Error("The entry key after the cache tail should have been 1")
	}

	cache.Delete("1")

	// (tail) 3 (head)
	if cache.tail.Key != "3" {
		t.Error("cache tail should have been the entry with key 3")
	}
	if cache.head.Key != "3" {
		t.Error("cache head should have been the entry with key 3")
	}

	if cache.head != cache.tail {
		t.Error("There should only be one entry in the cache")
	}
	if cache.head.previous != nil || cache.tail.next != nil {
		t.Error("Since head == tail, there should be no prev/next")
	}
}

func TestCache_SaveToFile(t *testing.T) {
	defer os.Remove(TestCacheFile)
	cache := NewCache()
	for n := 0; n < 10; n++ {
		cache.Set(strconv.Itoa(n), fmt.Sprintf("v%d", n))
	}
	err := cache.SaveToFile(TestCacheFile)
	if err != nil {
		t.Fatal("shouldn't have returned an error, but got:", err.Error())
	}
	newCache := NewCache()
	numberOfEntriesEvicted, err := newCache.ReadFromFile(TestCacheFile)
	if err != nil {
		t.Fatal("shouldn't have returned an error, but got:", err.Error())
	}
	if numberOfEntriesEvicted != 0 {
		t.Error("expected 0 entries to have been evicted, but got", numberOfEntriesEvicted)
	}
	if newCache.Count() != 10 {
		t.Error("expected newCache to have 10 entries, but got", newCache.Count())
	}
}

func TestCache_ReadFromFile(t *testing.T) {
	defer os.Remove(TestCacheFile)
	cache := NewCache()
	for n := 0; n < 10; n++ {
		cache.Set(strconv.Itoa(n), fmt.Sprintf("v%d", n))
	}
	err := cache.SaveToFile(TestCacheFile)
	if err != nil {
		panic(err)
	}
	newCache := cache.WithMaxSize(7)
	numberOfEntriesEvicted, err := newCache.ReadFromFile(TestCacheFile)
	if err != nil {
		panic(err)
	}
	if numberOfEntriesEvicted != 3 {
		t.Error("expected 3 entries to have been evicted, but got", numberOfEntriesEvicted)
	}
	if newCache.Count() != 7 {
		t.Error("expected newCache to have 7 entries since its MaxSize is 7, but got", newCache.Count())
	}
	for key, value := range cache.entries {
		expectedValue := fmt.Sprintf("v%s", key)
		if value.Value != expectedValue {
			t.Errorf("key %s should've had value '%s', but had '%s' instead", key, expectedValue, value.Value)
		}
	}
}
