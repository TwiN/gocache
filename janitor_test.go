package gocache

import (
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
