package gocache

import (
	"fmt"
	"testing"
	"time"
)

func TestCache_StartJanitor(t *testing.T) {
	cache := NewCache()
	cache.SetWithTTL("1", "1", time.Nanosecond)
	if cacheSize := cache.Count(); cacheSize != 1 {
		t.Errorf("expected cacheSize to be 1, but was %d", cacheSize)
	}
	err := cache.StartJanitor()
	if err != nil {
		t.Fatal(err)
	}
	defer cache.StopJanitor()
	time.Sleep(JanitorMinShiftBackOff * 2)
	if cacheSize := cache.Count(); cacheSize != 0 {
		t.Errorf("expected cacheSize to be 0, but was %d", cacheSize)
	}
}

func TestCache_StartJanitorWhenAlreadyStarted(t *testing.T) {
	cache := NewCache()
	if err := cache.StartJanitor(); err != nil {
		t.Fatal(err)
	}
	if err := cache.StartJanitor(); err == nil {
		t.Fatal("expected StartJanitor to return an error, because the janitor is already started")
	}
	cache.StopJanitor()
}

func TestJanitor(t *testing.T) {
	Debug = true
	cache := NewCache().WithMaxSize(3 * JanitorMaxIterationsPerShift)
	defer cache.Clear()
	for i := 0; i < 3*JanitorMaxIterationsPerShift; i++ {
		if i < JanitorMaxIterationsPerShift && i%2 == 0 {
			cache.SetWithTTL(fmt.Sprintf("%d", i), "value", time.Millisecond)
		} else {
			cache.SetWithTTL(fmt.Sprintf("%d", i), "value", time.Hour)
		}
	}
	cacheSize := cache.Count()
	err := cache.StartJanitor()
	if err != nil {
		t.Fatal(err)
	}
	defer cache.StopJanitor()
	time.Sleep(JanitorMinShiftBackOff * 4)
	if cacheSize <= cache.Count() {
		t.Error("The janitor should be deleting expired cache entries")
	}
	cacheSize = cache.Count()
	time.Sleep(JanitorMinShiftBackOff * 4)
	if cacheSize <= cache.Count() {
		t.Error("The janitor should be deleting expired cache entries")
	}
	cacheSize = cache.Count()
	time.Sleep(JanitorMinShiftBackOff * 4)
	if cacheSize <= cache.Count() {
		t.Error("The janitor should be deleting expired cache entries")
	}
	Debug = false
}

func TestJanitorIsLoopingProperly(t *testing.T) {
	cache := NewCache().WithMaxSize(JanitorMaxIterationsPerShift + 3)
	defer cache.Clear()
	for i := 0; i < JanitorMaxIterationsPerShift; i++ {
		cache.SetWithTTL(fmt.Sprintf("%d", i), "value", time.Hour)
	}
	cache.SetWithTTL("key-to-expire-1", "value", JanitorMinShiftBackOff*2)
	cache.SetWithTTL("key-to-expire-2", "value", JanitorMinShiftBackOff*2)
	cache.SetWithTTL("key-to-expire-3", "value", JanitorMinShiftBackOff*2)
	err := cache.StartJanitor()
	if err != nil {
		t.Fatal(err)
	}
	defer cache.StopJanitor()
	if cache.Count() != JanitorMaxIterationsPerShift+3 {
		t.Error("The janitor shouldn't have had enough time to remove anything from the cache yet", cache.Count())
	}
	time.Sleep(JanitorMinShiftBackOff * 8)
	if cache.Count() != JanitorMaxIterationsPerShift {
		t.Error("The janitor should've deleted 3 entries")
	}
}
