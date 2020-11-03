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
}

func TestJanitorIsLoopingProperly(t *testing.T) {
	cache := NewCache().WithMaxSize(3 * JanitorMaxIterationsPerShift)
	defer cache.Clear()
	cache.SetWithTTL("1", "value", time.Hour)
	cache.SetWithTTL("2", "value", JanitorMinShiftBackOff*3)
	cache.SetWithTTL("3", "value", JanitorMinShiftBackOff*3)
	cache.SetWithTTL("4", "value", JanitorMinShiftBackOff*3)
	cache.SetWithTTL("5", "value", time.Hour)
	err := cache.StartJanitor()
	if err != nil {
		t.Fatal(err)
	}
	defer cache.StopJanitor()
	if cache.Count() != 5 {
		t.Error("The janitor shouldn't have had enough time to remove anything from the cache yet")
	}
	time.Sleep(JanitorMinShiftBackOff * 4)
	if cache.Count() != 2 {
		t.Error("The janitor should've deleted 3 of the 5 entries")
	}
}
