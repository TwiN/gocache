package gocache

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"
)

const (
	TestCacheFile = "test.cache"
)

func TestNewCache(t *testing.T) {
	cache := NewCache().WithMaxSize(1234).WithEvictionPolicy(LeastRecentlyUsed)
	if cache.MaxMemoryUsage() != NoMaxMemoryUsage {
		t.Error("shouldn't have a max memory usage configured")
	}
	if cache.EvictionPolicy() != LeastRecentlyUsed {
		t.Error("should've had a LeastRecentlyUsed eviction policy")
	}
	if cache.MaxSize() != 1234 {
		t.Error("should've had a max cache size of 1234")
	}
	if cache.MemoryUsage() != 0 {
		t.Error("should've had a memory usage of 0")
	}
}

func TestCache_Stats(t *testing.T) {
	cache := NewCache().WithMaxSize(1234).WithEvictionPolicy(LeastRecentlyUsed)
	cache.Set("key", "value")
	if cache.Stats().Hits != 0 {
		t.Error("should have 0 hits")
	}
	if cache.Stats().Misses != 0 {
		t.Error("should have 0 misses")
	}
	cache.Get("key")
	if cache.Stats().Hits != 1 {
		t.Error("should have 1 hit")
	}
	if cache.Stats().Misses != 0 {
		t.Error("should have 0 misses")
	}
	cache.Get("key-that-does-not-exist")
	if cache.Stats().Hits != 1 {
		t.Error("should have 1 hit")
	}
	if cache.Stats().Misses != 1 {
		t.Error("should have 1 miss")
	}
}

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

func TestCache_GetExpired(t *testing.T) {
	cache := NewCache()
	cache.SetWithTTL("key", "value", time.Millisecond)
	time.Sleep(2 * time.Millisecond)
	_, ok := cache.Get("key")
	if ok {
		t.Error("expected key to be expired")
	}
}

func TestCache_GetEntryThatHasNotExpiredYet(t *testing.T) {
	cache := NewCache()
	cache.SetWithTTL("key", "value", time.Hour)
	_, ok := cache.Get("key")
	if !ok {
		t.Error("expected key to not have expired")
	}
}

func TestCache_GetAll(t *testing.T) {
	cache := NewCache().WithMaxSize(10)
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	keyValues := cache.GetAll([]string{"key1", "key2", "key3"})
	if len(keyValues) != 3 {
		t.Error("expected length of map to be 3")
	}
	if keyValues["key1"] != "value1" {
		t.Errorf("expected: %s, but got: %s", "value1", keyValues["key1"])
	}
	if keyValues["key2"] != "value2" {
		t.Errorf("expected: %s, but got: %s", "value2", keyValues["key2"])
	}
	if value, ok := keyValues["key3"]; !ok || value != nil {
		t.Errorf("expected key3 to exist and be nil, but got: %s", value)
	}
}

func TestCache_GetKeysByPattern(t *testing.T) {
	// All keys match
	testGetKeysByPattern(t, []string{"key1", "key2", "key3", "key4"}, "key*", 4)
	testGetKeysByPattern(t, []string{"key1", "key2", "key3", "key4"}, "*y*", 4)
	testGetKeysByPattern(t, []string{"key1", "key2", "key3", "key4"}, "*key*", 4)
	testGetKeysByPattern(t, []string{"key1", "key2", "key3", "key4"}, "*", 4)
	// Some keys match
	testGetKeysByPattern(t, []string{"key1", "key2", "key3", "key4", "key11"}, "key1*", 2)
	testGetKeysByPattern(t, []string{"key1", "key2", "key3", "key4", "key11"}, "key1*", 2)
	testGetKeysByPattern(t, []string{"key1", "key2", "key3", "key4", "key11", "key111"}, "key1*", 3)
	testGetKeysByPattern(t, []string{"key1", "key2", "key3", "key4", "key11", "key111"}, "key11*", 2)
	testGetKeysByPattern(t, []string{"key1", "key2", "key3", "key4", "key11", "key111"}, "*11*", 2)
	testGetKeysByPattern(t, []string{"key1", "key2", "key3", "key4", "key11", "key111"}, "k*1*", 3)
	testGetKeysByPattern(t, []string{"key1", "key2", "key3", "key4", "key11", "key111"}, "*k*1", 3)
	// No keys match
	testGetKeysByPattern(t, []string{"key1", "key2", "key3", "key4"}, "image*", 0)
	testGetKeysByPattern(t, []string{"key1", "key2", "key3", "key4"}, "?", 0)
}

func testGetKeysByPattern(t *testing.T, keys []string, pattern string, expectedMatchingKeys int) {
	cache := NewCache().WithMaxSize(len(keys))
	for _, key := range keys {
		cache.Set(key, key)
	}
	matchingKeys := cache.GetKeysByPattern(pattern, 0)
	if len(matchingKeys) != expectedMatchingKeys {
		t.Errorf("expected to have %d keys to match pattern '%s', got %d", expectedMatchingKeys, pattern, len(matchingKeys))
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

func TestCache_SetDifferentTypesOfData(t *testing.T) {
	cache := NewCache().WithMaxSize(10)
	cache.Set("key", 1)
	value, ok := cache.Get("key")
	if !ok {
		t.Error("expected key to exist")
	}
	if value != 1 {
		t.Errorf("expected: %v, but got: %v", 1, value)
	}
	cache.Set("key", struct{ Test string }{Test: "test"})
	value, ok = cache.Get("key")
	if !ok {
		t.Error("expected key to exist")
	}
	if value.(struct{ Test string }) != struct{ Test string }{Test: "test"} {
		t.Errorf("expected: %s, but got: %s", "newvalue", value)
	}
}

func TestCache_SetGetIntValue(t *testing.T) {
	cache := NewCache().WithMaxSize(10)
	cache.Set("key", 1)
	value, ok := cache.Get("key")
	if !ok {
		t.Error("expected key to exist")
	}
	if value != 1 {
		t.Errorf("expected: %v, but got: %v", 1, value)
	}
	cache.Set("key", 2.1)
	value, ok = cache.Get("key")
	if !ok {
		t.Error("expected key to exist")
	}
	if value != 2.1 {
		t.Errorf("expected: %v, but got: %v", 2.1, value)
	}
}

func TestCache_SetGetBool(t *testing.T) {
	cache := NewCache().WithMaxSize(10)
	cache.Set("key", true)
	value, ok := cache.Get("key")
	if !ok {
		t.Error("expected key to exist")
	}
	if value != true {
		t.Errorf("expected: %v, but got: %v", true, value)
	}
}

func TestCache_SetGetByteSlice(t *testing.T) {
	cache := NewCache().WithMaxSize(10)
	cache.Set("key", []byte("hey"))
	value, ok := cache.Get("key")
	if !ok {
		t.Error("expected key to exist")
	}
	if bytes.Compare(value.([]byte), []byte("hey")) != 0 {
		t.Errorf("expected: %v, but got: %v", []byte("hey"), value)
	}
}

func TestCache_SetGetStringSlice(t *testing.T) {
	cache := NewCache().WithMaxSize(10)
	cache.Set("key", []string{"john", "doe"})
	value, ok := cache.Get("key")
	if !ok {
		t.Error("expected key to exist")
	}
	if value.([]string)[0] != "john" {
		t.Errorf("expected: %v, but got: %v", "john", value)
	}
	if value.([]string)[1] != "doe" {
		t.Errorf("expected: %v, but got: %v", "doe", value)
	}
}

func TestCache_SetGetStruct(t *testing.T) {
	cache := NewCache().WithMaxSize(10)
	type Custom struct {
		Int     int
		Uint    uint
		Float32 float32
		String  string
		Strings []string
		Nested  struct {
			String string
		}
	}
	cache.Set("key", Custom{
		Int:     111,
		Uint:    222,
		Float32: 123.456,
		String:  "hello",
		Strings: []string{"s1", "s2"},
		Nested:  struct{ String string }{String: "nested field"},
	})
	value, ok := cache.Get("key")
	if !ok {
		t.Error("expected key to exist")
	}
	if ExpectedValue := 111; value.(Custom).Int != ExpectedValue {
		t.Errorf("expected: %v, but got: %v", ExpectedValue, value)
	}
	if ExpectedValue := uint(222); value.(Custom).Uint != ExpectedValue {
		t.Errorf("expected: %v, but got: %v", ExpectedValue, value)
	}
	if ExpectedValue := float32(123.456); value.(Custom).Float32 != ExpectedValue {
		t.Errorf("expected: %v, but got: %v", ExpectedValue, value)
	}
	if ExpectedValue := "hello"; value.(Custom).String != ExpectedValue {
		t.Errorf("expected: %v, but got: %v", ExpectedValue, value)
	}
	if ExpectedValue := "s1"; value.(Custom).Strings[0] != ExpectedValue {
		t.Errorf("expected: %v, but got: %v", ExpectedValue, value)
	}
	if ExpectedValue := "s2"; value.(Custom).Strings[1] != ExpectedValue {
		t.Errorf("expected: %v, but got: %v", ExpectedValue, value)
	}
	if ExpectedValue := "nested field"; value.(Custom).Nested.String != ExpectedValue {
		t.Errorf("expected: %v, but got: %v", ExpectedValue, value)
	}
}

func TestCache_SetAll(t *testing.T) {
	cache := NewCache().WithMaxSize(10)
	cache.SetAll(map[string]interface{}{"k1": "v1", "k2": "v2"})
	value, ok := cache.Get("k1")
	if !ok {
		t.Error("expected key to exist")
	}
	if value != "v1" {
		t.Errorf("expected: %s, but got: %s", "v1", value)
	}
	value, ok = cache.Get("k2")
	if !ok {
		t.Error("expected key to exist")
	}
	if value != "v2" {
		t.Errorf("expected: %s, but got: %s", "v2", value)
	}
	cache.SetAll(map[string]interface{}{"k1": "updated"})
	value, ok = cache.Get("k1")
	if !ok {
		t.Error("expected key to exist")
	}
	if value != "updated" {
		t.Errorf("expected: %s, but got: %s", "updated", value)
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
	cache.Set("2", []byte("value"))
	cache.Set("3", []byte("value"))
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
	cache.Set("2", []byte("value"))
	cache.Set("3", []byte("value"))
	_, _ = cache.Get("1")
	cache.Set("4", []byte("value"))

	_, ok := cache.Get("1")
	if !ok {
		t.Error("expected key 1 to still exist, because LRU")
	}
}

func TestCache_HeadToTailSimple(t *testing.T) {
	cache := NewCache().WithMaxSize(3)
	cache.Set("1", "1")
	if cache.tail.Key != "1" && cache.head.Key != "1" {
		t.Error("expected tail=1 and head=1")
	}
	cache.Set("2", "2")
	if cache.tail.Key != "1" && cache.head.Key != "2" {
		t.Error("expected tail=1 and head=2")
	}
	cache.Set("3", "3")
	if cache.tail.Key != "1" && cache.head.Key != "3" {
		t.Error("expected tail=1 and head=4")
	}
	cache.Set("4", "4")
	if cache.tail.Key != "2" && cache.head.Key != "4" {
		t.Error("expected tail=2 and head=4")
	}
	cache.Set("5", "5")
	if cache.tail.Key != "3" && cache.head.Key != "5" {
		t.Error("expected tail=3 and head=5")
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

func TestCache_HeadStaysTheSameIfCallRepeatedly(t *testing.T) {
	cache := NewCache().WithEvictionPolicy(LeastRecentlyUsed).WithMaxSize(10)
	cache.Set("1", "1")
	if cache.tail.Key != "1" && cache.head.Key != "1" {
		t.Error("expected tail=1 and head=1")
	}
	cache.Set("1", "1")
	if cache.tail.Key != "1" && cache.head.Key != "1" {
		t.Error("expected tail=1 and head=1")
	}
	cache.Get("1")
	if cache.tail.Key != "1" && cache.head.Key != "1" {
		t.Error("expected tail=1 and head=1")
	}
	cache.Get("1")
	if cache.tail.Key != "1" && cache.head.Key != "1" {
		t.Error("expected tail=1 and head=1")
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

	cache.Set("1", "hey")
	cache.Set("2", []byte("sup"))
	cache.Set("3", 123456)

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

func TestCache_DeleteAll(t *testing.T) {
	cache := NewCache()
	cache.Set("1", []byte("1"))
	cache.Set("2", []byte("2"))
	cache.Set("3", []byte("3"))
	if len(cache.GetAll([]string{"1", "2", "3"})) != 3 {
		t.Error("Expected keys 1, 2 and 3 to exist")
	}
	numberOfDeletedKeys := cache.DeleteAll([]string{"1", "2", "3"})
	if numberOfDeletedKeys != 3 {
		t.Errorf("Expected 3 keys to have been deleted, but only %d were deleted", numberOfDeletedKeys)
	}
}

func TestCache_TTL(t *testing.T) {
	cache := NewCache()
	ttl, err := cache.TTL("key")
	if err != ErrKeyDoesNotExist {
		t.Errorf("expected %s, got %s", ErrKeyDoesNotExist, err)
	}
	cache.Set("key", "value")
	_, err = cache.TTL("key")
	if err != ErrKeyHasNoExpiration {
		t.Error("Expected TTL on new key created using Set to have no expiration")
	}
	cache.SetWithTTL("key", "value", time.Hour)
	ttl, err = cache.TTL("key")
	if err != nil {
		t.Error("Unexpected error")
	}
	if ttl.Minutes() < 59 || ttl.Minutes() > 60 {
		t.Error("Expected the TTL to be almost an hour")
	}
	cache.SetWithTTL("key", "value", 5*time.Millisecond)
	time.Sleep(6 * time.Millisecond)
	ttl, err = cache.TTL("key")
	if err != ErrKeyDoesNotExist {
		t.Error("key should've expired, thus TTL should've returned ")
	}
}

func TestCache_Expire(t *testing.T) {
	cache := NewCache()
	if cache.Expire("key-that-does-not-exist", time.Minute) {
		t.Error("Expected Expire to return false, because the key used did not exist")
	}
	cache.Set("key", "value")
	_, err := cache.TTL("key")
	if err != ErrKeyHasNoExpiration {
		t.Error("Expected TTL on new key created using Set to have no expiration")
	}
	if !cache.Expire("key", time.Hour) {
		t.Error("Expected Expire to return true")
	}
	ttl, err := cache.TTL("key")
	if err != nil {
		t.Error("Unexpected error")
	}
	if ttl.Minutes() < 59 || ttl.Minutes() > 60 {
		t.Error("Expected the TTL to be almost an hour")
	}
	if !cache.Expire("key", 5*time.Millisecond) {
		t.Error("Expected Expire to return true")
	}
	time.Sleep(6 * time.Millisecond)
	_, err = cache.TTL("key")
	if err != ErrKeyDoesNotExist {
		t.Error("key should've expired, thus TTL should've returned ")
	}
	if cache.Expire("key", time.Hour) {
		t.Error("Expire should've returned false, because the key should've already expired, thus no longer exist")
	}
}

func TestCache_Clear(t *testing.T) {
	cache := NewCache().WithMaxSize(10)
	cache.Set("k1", "v1")
	cache.Set("k2", "v2")
	cache.Set("k3", "v3")
	if cache.Count() != 3 {
		t.Error("expected cache size to be 3, got", cache.Count())
	}
	cache.Clear()
	if cache.Count() != 0 {
		t.Error("expected cache to be empty")
	}
	if cache.memoryUsage != 0 {
		t.Error("expected cache.memoryUsage to be 0")
	}
}

func TestCache_WithMaxMemoryUsage(t *testing.T) {
	const ValueSize = Kilobyte
	cache := NewCache().WithMaxSize(0).WithMaxMemoryUsage(Kilobyte * 64)
	for i := 0; i < 100; i++ {
		cache.Set(fmt.Sprintf("%d", i), strings.Repeat("0", ValueSize))
	}
	if cache.MemoryUsage()/1024 < 63 || cache.MemoryUsage()/1024 > 65 {
		t.Error("expected memoryUsage to be between 63KB and 64KB")
	}
}

func TestCache_WithMaxMemoryUsageWhenAddingAnEntryThatCausesMoreThanOneEviction(t *testing.T) {
	const ValueSize = Kilobyte
	cache := NewCache().WithMaxSize(0).WithMaxMemoryUsage(64 * Kilobyte)
	for i := 0; i < 100; i++ {
		cache.Set(fmt.Sprintf("%d", i), strings.Repeat("0", ValueSize))
	}
	if cache.MemoryUsage()/1024 < 63 || cache.MemoryUsage()/1024 > 65 {
		t.Error("expected memoryUsage to be between 63KB and 64KB")
	}
}

func TestCache_memoryUsageAfterSet10000AndDelete5000(t *testing.T) {
	const ValueSize = 64
	cache := NewCache().WithMaxSize(10000).WithMaxMemoryUsage(Gigabyte)
	for i := 0; i < cache.maxSize; i++ {
		cache.Set(fmt.Sprintf("%05d", i), strings.Repeat("0", ValueSize))
	}
	memoryUsageBeforeDeleting := cache.MemoryUsage()
	for i := 0; i < cache.maxSize/2; i++ {
		key := fmt.Sprintf("%05d", i)
		cache.Delete(key)
	}
	memoryUsageRatio := float32(cache.MemoryUsage()) / float32(memoryUsageBeforeDeleting)
	if memoryUsageRatio != 0.5 {
		t.Error("Since half of the keys were deleted, the memoryUsage should've been half of what the memory usage was before beginning deletion")
	}
}

func TestCache_memoryUsageIsReliable(t *testing.T) {
	cache := NewCache().WithMaxMemoryUsage(Megabyte)
	previousCacheMemoryUsage := cache.MemoryUsage()
	if previousCacheMemoryUsage != 0 {
		t.Error("cache.MemoryUsage() should've been 0")
	}
	cache.Set("1", 1)
	if cache.MemoryUsage() <= previousCacheMemoryUsage {
		t.Error("cache.MemoryUsage() should've increased")
	}
	previousCacheMemoryUsage = cache.MemoryUsage()
	cache.SetAll(map[string]interface{}{"2": "2", "3": "3", "4": "4"})
	if cache.MemoryUsage() <= previousCacheMemoryUsage {
		t.Error("cache.MemoryUsage() should've increased")
	}
	previousCacheMemoryUsage = cache.MemoryUsage()
	cache.Delete("2")
	if cache.MemoryUsage() >= previousCacheMemoryUsage {
		t.Error("cache.MemoryUsage() should've decreased")
	}
	previousCacheMemoryUsage = cache.MemoryUsage()
	cache.Set("1", 1)
	if cache.MemoryUsage() != previousCacheMemoryUsage {
		t.Error("cache.MemoryUsage() shouldn't have changed, because the entry didn't change")
	}
	previousCacheMemoryUsage = cache.MemoryUsage()
	cache.Delete("3")
	if cache.MemoryUsage() >= previousCacheMemoryUsage {
		t.Error("cache.MemoryUsage() should've decreased")
	}
	previousCacheMemoryUsage = cache.MemoryUsage()
	cache.Delete("4")
	if cache.MemoryUsage() >= previousCacheMemoryUsage {
		t.Error("cache.MemoryUsage() should've decreased")
	}
	previousCacheMemoryUsage = cache.MemoryUsage()
	cache.Delete("1")
	if cache.MemoryUsage() >= previousCacheMemoryUsage || cache.memoryUsage != 0 {
		t.Error("cache.MemoryUsage() should've been 0")
	}
	previousCacheMemoryUsage = cache.MemoryUsage()
	cache.Set("1", "v4lu3")
	if cache.MemoryUsage() <= previousCacheMemoryUsage {
		t.Error("cache.MemoryUsage() should've increased")
	}
	previousCacheMemoryUsage = cache.MemoryUsage()
	cache.Set("1", "value")
	if cache.MemoryUsage() != previousCacheMemoryUsage {
		t.Error("cache.MemoryUsage() shouldn't have changed")
	}
	previousCacheMemoryUsage = cache.MemoryUsage()
	cache.Set("1", true)
	if cache.MemoryUsage() >= previousCacheMemoryUsage {
		t.Error("cache.MemoryUsage() should've decreased, because a bool uses less memory than a string")
	}
}

func TestEvictionWhenThereIsNothingToEvict(t *testing.T) {
	cache := NewCache()
	cache.evict()
	cache.evict()
	cache.evict()
}
