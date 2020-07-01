package gocache

import (
	"fmt"
	"testing"
	"time"
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