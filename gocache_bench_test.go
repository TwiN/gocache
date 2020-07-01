package gocache

import (
	"math/rand"
	"strconv"
	"strings"
	"testing"
)

func BenchmarkMap_Get(b *testing.B) {
	m := make(map[string]interface{})
	for n := 0; n < b.N; n++ {
		_, _ = m[strconv.Itoa(n)]
	}
}

func BenchmarkMap_SetSmallValue(b *testing.B) {
	value := "a"
	m := make(map[string]interface{})
	for n := 0; n < b.N; n++ {
		m[strconv.Itoa(n)] = &value
	}
}

func BenchmarkMap_SetMediumValue(b *testing.B) {
	value := strings.Repeat("a", 1024)
	m := make(map[string]interface{})
	for n := 0; n < b.N; n++ {
		m[strconv.Itoa(n)] = &value
	}
}

func BenchmarkMap_SetLargeValue(b *testing.B) {
	value := strings.Repeat("a", 1024*100)
	m := make(map[string]interface{})
	for n := 0; n < b.N; n++ {
		m[strconv.Itoa(n)] = &value
	}
}

func BenchmarkCache_SetSmallValueWithMaxSize10(b *testing.B) {
	value := "a"
	cache := NewCache().WithMaxSize(10)
	writeToCache(b, cache, value)
}

func BenchmarkCache_SetMediumValueWithMaxSize10(b *testing.B) {
	value := strings.Repeat("a", 1024)
	cache := NewCache().WithMaxSize(10)
	writeToCache(b, cache, value)
}

func BenchmarkCache_SetLargeValueWithMaxSize10(b *testing.B) {
	value := strings.Repeat("a", 1024*100)
	cache := NewCache().WithMaxSize(10)
	writeToCache(b, cache, value)
}

func BenchmarkCache_SetSmallValueWithMaxSize1000(b *testing.B) {
	value := "a"
	cache := NewCache().WithMaxSize(1000)
	writeToCache(b, cache, value)
}

func BenchmarkCache_SetMediumValueWithMaxSize1000(b *testing.B) {
	value := strings.Repeat("a", 1024)
	cache := NewCache().WithMaxSize(1000)
	writeToCache(b, cache, value)
}

func BenchmarkCache_SetLargeValueWithMaxSize1000(b *testing.B) {
	value := strings.Repeat("a", 1024*100)
	cache := NewCache().WithMaxSize(1000)
	writeToCache(b, cache, value)
}

func BenchmarkCache_SetSmallValueWithMaxSize100000(b *testing.B) {
	value := "a"
	cache := NewCache().WithMaxSize(100000)
	writeToCache(b, cache, value)
}

func BenchmarkCache_SetMediumValueWithMaxSize100000(b *testing.B) {
	value := strings.Repeat("a", 1024)
	cache := NewCache().WithMaxSize(100000)
	writeToCache(b, cache, value)
}

func BenchmarkCache_SetLargeValueWithMaxSize100000(b *testing.B) {
	value := strings.Repeat("a", 1024*100)
	cache := NewCache().WithMaxSize(100000)
	writeToCache(b, cache, value)
}

func BenchmarkCache_SetSmallValueWithMaxSize100000AndLRU(b *testing.B) {
	value := "a"
	cache := NewCache().WithMaxSize(100000).WithEvictionPolicy(LeastRecentlyUsed)
	writeToCache(b, cache, value)
}

func BenchmarkCache_SetMediumValueWithMaxSize100000AndLRU(b *testing.B) {
	value := strings.Repeat("a", 1024)
	cache := NewCache().WithMaxSize(100000).WithEvictionPolicy(LeastRecentlyUsed)
	writeToCache(b, cache, value)
}

func BenchmarkCache_SetLargeValueWithMaxSize100000AndLRU(b *testing.B) {
	value := strings.Repeat("a", 1024*100)
	cache := NewCache().WithMaxSize(100000).WithEvictionPolicy(LeastRecentlyUsed)
	writeToCache(b, cache, value)
}

func writeToCache(b *testing.B, cache *Cache, value string) {
	for n := 0; n < b.N; n++ {
		cache.Set(strconv.Itoa(n), value)
	}
}

func BenchmarkCache_GetWithMaxSize10(b *testing.B) {
	cache := NewCache().WithMaxSize(10)

	for n := 0; n < b.N; n++ {
		cache.Get(strconv.Itoa(n))
	}
}

func BenchmarkCache_GetAndSetConcurrentlyWithMaxSize10(b *testing.B) {
	data := map[string]string{
		"k1": "v1",
		"k2": "v2",
		"k3": "v3",
		"k4": "v4",
	}
	cache := NewCache().WithMaxSize(10)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for k, v := range data {
				cache.Set(k, v)
				val, ok := cache.Get(k)
				if !ok {
					b.Errorf("key: %v; value: %v", k, v)
				}
				if v != val {
					b.Errorf("expected: %v; got: %v", v, val)
				}
			}
		}
	})
}

func BenchmarkCache_GetConcurrently(b *testing.B) {
	testValue := strings.Repeat("a", 256)
	cache := NewCache().WithMaxSize(b.N)
	for i := 0; i < b.N; i++ {
		cache.Set(strconv.Itoa(i), testValue)
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key := strconv.Itoa(rand.Intn(b.N))
			val, ok := cache.Get(key)
			if !ok {
				b.Errorf("key: %v; value: %v", key, val)
			}
			if val != testValue {
				b.Errorf("expected: %v; got: %v", val, testValue)
			}
		}
	})
}

func BenchmarkCache_GetKeysThatDoNotExistConcurrently(b *testing.B) {
	cache := NewCache().WithMaxSize(1000)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, ok := cache.Get(strconv.Itoa(rand.Intn(b.N))); ok {
				b.Errorf("Cache should've been empty")
			}
		}
	})
}
