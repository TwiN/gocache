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
		m[strconv.Itoa(n)] = value
	}
	b.ReportAllocs()
}

func BenchmarkMap_SetMediumValue(b *testing.B) {
	value := strings.Repeat("a", 1024)
	m := make(map[string]interface{})
	for n := 0; n < b.N; n++ {
		m[strconv.Itoa(n)] = value
	}
	b.ReportAllocs()
}

func BenchmarkMap_SetLargeValue(b *testing.B) {
	value := strings.Repeat("a", 1024*100)
	m := make(map[string]interface{})
	for n := 0; n < b.N; n++ {
		m[strconv.Itoa(n)] = value
	}
	b.ReportAllocs()
}

//func BenchmarkPatrickmnGoCache_SetSmallValue(b *testing.B) {
//	value := "a"
//	cache := patrickmnGoCache.New(0, 0)
//	for n := 0; n < b.N; n++ {
//		cache.Set(strconv.Itoa(n), value, 0)
//	}
//}
//
//func BenchmarkPatrickmnGoCache_SetMediumValue(b *testing.B) {
//	value := strings.Repeat("a", 1024)
//	cache := patrickmnGoCache.New(0, 0)
//	for n := 0; n < b.N; n++ {
//		cache.Set(strconv.Itoa(n), value, 0)
//	}
//}
//
//func BenchmarkPatrickmnGoCache_SetLargeValue(b *testing.B) {
//	value := strings.Repeat("a", 1024*100)
//	cache := patrickmnGoCache.New(0, 0)
//	for n := 0; n < b.N; n++ {
//		cache.Set(strconv.Itoa(n), value, 0)
//	}
//}

func BenchmarkCache_Get(b *testing.B) {
	cache := NewCache().WithMaxSize(NoMaxSize).WithMaxMemoryUsage(NoMaxMemoryUsage)
	for n := 0; n < b.N; n++ {
		cache.Get(strconv.Itoa(n))
	}
	b.ReportAllocs()
}

func BenchmarkCache_SetSmallValue(b *testing.B) {
	value := "a"
	cache := NewCache().WithMaxSize(NoMaxSize).WithMaxMemoryUsage(NoMaxMemoryUsage)
	for n := 0; n < b.N; n++ {
		cache.SetWithTTL(strconv.Itoa(n), value, 0)
	}
	b.ReportAllocs()
}

func BenchmarkCache_SetMediumValue(b *testing.B) {
	value := strings.Repeat("a", 1024)
	cache := NewCache().WithMaxSize(NoMaxSize).WithMaxMemoryUsage(NoMaxMemoryUsage)
	for n := 0; n < b.N; n++ {
		cache.Set(strconv.Itoa(n), value)
	}
	b.ReportAllocs()
}

func BenchmarkCache_SetLargeValue(b *testing.B) {
	value := strings.Repeat("a", 1024*100)
	cache := NewCache().WithMaxSize(NoMaxSize).WithMaxMemoryUsage(NoMaxMemoryUsage)
	for n := 0; n < b.N; n++ {
		cache.Set(strconv.Itoa(n), value)
	}
	b.ReportAllocs()
}

func BenchmarkCache_GetUsingLRU(b *testing.B) {
	cache := NewCache().WithMaxSize(NoMaxSize).WithMaxMemoryUsage(NoMaxMemoryUsage).WithEvictionPolicy(LeastRecentlyUsed)
	for n := 0; n < b.N; n++ {
		cache.Get(strconv.Itoa(n))
	}
	b.ReportAllocs()
}

func BenchmarkCache_SetSmallValueUsingLRU(b *testing.B) {
	value := "a"
	cache := NewCache().WithMaxSize(NoMaxSize).WithMaxMemoryUsage(NoMaxMemoryUsage).WithEvictionPolicy(LeastRecentlyUsed)
	for n := 0; n < b.N; n++ {
		cache.Set(strconv.Itoa(n), value)
	}
	b.ReportAllocs()
}

func BenchmarkCache_SetMediumValueUsingLRU(b *testing.B) {
	value := strings.Repeat("a", 1024)
	cache := NewCache().WithMaxSize(NoMaxSize).WithMaxMemoryUsage(NoMaxMemoryUsage).WithEvictionPolicy(LeastRecentlyUsed)
	for n := 0; n < b.N; n++ {
		cache.Set(strconv.Itoa(n), value)
	}
	b.ReportAllocs()
}

func BenchmarkCache_SetLargeValueUsingLRU(b *testing.B) {
	value := strings.Repeat("a", 1024*100)
	cache := NewCache().WithMaxSize(NoMaxSize).WithMaxMemoryUsage(NoMaxMemoryUsage).WithEvictionPolicy(LeastRecentlyUsed)
	for n := 0; n < b.N; n++ {
		cache.Set(strconv.Itoa(n), value)
	}
	b.ReportAllocs()
}

func BenchmarkCache_SetSmallValueWhenUsingMaxMemoryUsage(b *testing.B) {
	value := "a"
	cache := NewCache().WithMaxSize(NoMaxSize).WithMaxMemoryUsage(999 * Gigabyte)
	for n := 0; n < b.N; n++ {
		cache.Set(strconv.Itoa(n), value)
	}
	b.ReportAllocs()
}

func BenchmarkCache_SetMediumValueWhenUsingMaxMemoryUsage(b *testing.B) {
	value := strings.Repeat("a", 1024)
	cache := NewCache().WithMaxSize(NoMaxSize).WithMaxMemoryUsage(999 * Gigabyte)
	for n := 0; n < b.N; n++ {
		cache.Set(strconv.Itoa(n), value)
	}
	b.ReportAllocs()
}

func BenchmarkCache_SetLargeValueWhenUsingMaxMemoryUsage(b *testing.B) {
	value := strings.Repeat("a", 1024*100)
	cache := NewCache().WithMaxSize(NoMaxSize).WithMaxMemoryUsage(999 * Gigabyte)
	for n := 0; n < b.N; n++ {
		cache.Set(strconv.Itoa(n), value)
	}
	b.ReportAllocs()
}

// go test -cpuprofile cpu.prof -memprofile mem.prof -bench ^\QBenchmarkCache_SetSmallValueWhenUsingMaxMemoryUsage\E$
func BenchmarkCache_SetSmallValueWithMaxSize10(b *testing.B) {
	value := "a"
	cache := NewCache().WithMaxSize(10).WithMaxMemoryUsage(NoMaxMemoryUsage)
	writeToCache(b, cache, value)
}

func BenchmarkCache_SetMediumValueWithMaxSize10(b *testing.B) {
	value := strings.Repeat("a", 1024)
	cache := NewCache().WithMaxSize(10).WithMaxMemoryUsage(NoMaxMemoryUsage)
	writeToCache(b, cache, value)
}

func BenchmarkCache_SetLargeValueWithMaxSize10(b *testing.B) {
	value := strings.Repeat("a", 1024*100)
	cache := NewCache().WithMaxSize(10).WithMaxMemoryUsage(NoMaxMemoryUsage)
	writeToCache(b, cache, value)
}

func BenchmarkCache_SetSmallValueWithMaxSize1000(b *testing.B) {
	value := "a"
	cache := NewCache().WithMaxSize(1000).WithMaxMemoryUsage(NoMaxMemoryUsage)
	writeToCache(b, cache, value)
}

func BenchmarkCache_SetMediumValueWithMaxSize1000(b *testing.B) {
	value := strings.Repeat("a", 1024)
	cache := NewCache().WithMaxSize(1000).WithMaxMemoryUsage(NoMaxMemoryUsage)
	writeToCache(b, cache, value)
}

func BenchmarkCache_SetLargeValueWithMaxSize1000(b *testing.B) {
	value := strings.Repeat("a", 1024*100)
	cache := NewCache().WithMaxSize(1000).WithMaxMemoryUsage(NoMaxMemoryUsage)
	writeToCache(b, cache, value)
}

func BenchmarkCache_SetSmallValueWithMaxSize100000(b *testing.B) {
	value := "a"
	cache := NewCache().WithMaxSize(100000).WithMaxMemoryUsage(NoMaxMemoryUsage)
	writeToCache(b, cache, value)
}

func BenchmarkCache_SetMediumValueWithMaxSize100000(b *testing.B) {
	value := strings.Repeat("a", 1024)
	cache := NewCache().WithMaxSize(100000).WithMaxMemoryUsage(NoMaxMemoryUsage)
	writeToCache(b, cache, value)
}

func BenchmarkCache_SetLargeValueWithMaxSize100000(b *testing.B) {
	value := strings.Repeat("a", 1024*100)
	cache := NewCache().WithMaxSize(100000).WithMaxMemoryUsage(NoMaxMemoryUsage)
	writeToCache(b, cache, value)
}

func BenchmarkCache_SetSmallValueWithMaxSize100000AndLRU(b *testing.B) {
	value := "a"
	cache := NewCache().WithMaxSize(100000).WithMaxMemoryUsage(NoMaxMemoryUsage).WithEvictionPolicy(LeastRecentlyUsed)
	writeToCache(b, cache, value)
}

func BenchmarkCache_SetMediumValueWithMaxSize100000AndLRU(b *testing.B) {
	value := strings.Repeat("a", 1024)
	cache := NewCache().WithMaxSize(100000).WithMaxMemoryUsage(NoMaxMemoryUsage).WithEvictionPolicy(LeastRecentlyUsed)
	writeToCache(b, cache, value)
}

func BenchmarkCache_SetLargeValueWithMaxSize100000AndLRU(b *testing.B) {
	value := strings.Repeat("a", 1024*100)
	cache := NewCache().WithMaxSize(100000).WithMaxMemoryUsage(NoMaxMemoryUsage).WithEvictionPolicy(LeastRecentlyUsed)
	writeToCache(b, cache, value)
}

func writeToCache(b *testing.B, cache *Cache, value string) {
	for n := 0; n < b.N; n++ {
		cache.Set(strconv.Itoa(n), value)
	}
	b.ReportAllocs()
}

func BenchmarkCache_GetAndSetMultipleConcurrently(b *testing.B) {
	data := map[string]string{
		"k1": "v1",
		"k2": "v2",
		"k3": "v3",
		"k4": "v4",
		"k5": "v5",
		"k6": "v6",
		"k7": "v7",
		"k8": "v8",
	}
	cache := NewCache().WithMaxSize(NoMaxSize)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for k, v := range data {
				cache.Set(k, v)
				cache.Get(k)
			}
		}
	})
	b.ReportAllocs()
}

//func BenchmarkPatrickmnGoCache_GetAndSetMultipleConcurrently(b *testing.B) {
//	data := map[string]string{
//		"k1": "v1",
//		"k2": "v2",
//		"k3": "v3",
//		"k4": "v4",
//		"k5": "v5",
//		"k6": "v6",
//		"k7": "v7",
//		"k8": "v8",
//	}
//	cache := patrickmnGoCache.New(0, 0)
//
//	b.RunParallel(func(pb *testing.PB) {
//		for pb.Next() {
//			for k, v := range data {
//				cache.Set(k, v, 0)
//				cache.Get(k)
//			}
//		}
//	})
//}

func BenchmarkCache_GetAndSetConcurrentlyWithRandomKeysAndLRU(b *testing.B) {
	testValue := strings.Repeat("a", 256)
	cache := NewCache().WithEvictionPolicy(LeastRecentlyUsed).WithMaxSize(1000).WithMaxMemoryUsage(NoMaxMemoryUsage)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			k := strconv.Itoa(rand.Intn(b.N))
			cache.Set(k, testValue)
			_, _ = cache.Get(k)
		}
	})
	b.ReportAllocs()
}

func BenchmarkCache_GetAndSetConcurrentlyWithRandomKeysAndFIFO(b *testing.B) {
	testValue := strings.Repeat("a", 256)
	cache := NewCache().WithEvictionPolicy(FirstInFirstOut).WithMaxSize(1000).WithMaxMemoryUsage(NoMaxMemoryUsage)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			k := strconv.Itoa(rand.Intn(b.N))
			cache.Set(k, testValue)
			_, _ = cache.Get(k)
		}
	})
	b.ReportAllocs()
}

func BenchmarkCache_GetAndSetConcurrentlyWithRandomKeysAndNoEvictionAndLRU(b *testing.B) {
	testValue := strings.Repeat("a", 256)
	cache := NewCache().WithEvictionPolicy(LeastRecentlyUsed).WithMaxSize(b.N).WithMaxMemoryUsage(NoMaxMemoryUsage)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			k := strconv.Itoa(rand.Intn(b.N))
			cache.Set(k, testValue)
			_, _ = cache.Get(k)
		}
	})
	b.ReportAllocs()
}

func BenchmarkCache_GetAndSetConcurrentlyWithRandomKeysAndNoEvictionAndFIFO(b *testing.B) {
	testValue := strings.Repeat("a", 256)
	cache := NewCache().WithEvictionPolicy(FirstInFirstOut).WithMaxSize(b.N).WithMaxMemoryUsage(NoMaxMemoryUsage)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			k := strconv.Itoa(rand.Intn(b.N))
			cache.Set(k, testValue)
			_, _ = cache.Get(k)
		}
	})
	b.ReportAllocs()
}

func BenchmarkCache_GetAndSetConcurrentlyWithFrequentEvictionsAndLRU(b *testing.B) {
	testValue := strings.Repeat("a", 256)
	cache := NewCache().WithEvictionPolicy(LeastRecentlyUsed).WithMaxSize(3).WithMaxMemoryUsage(NoMaxMemoryUsage)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			k := strconv.Itoa(rand.Intn(15))
			cache.Set(k, testValue)
			_, _ = cache.Get(k)
		}
	})
	b.ReportAllocs()
}

func BenchmarkCache_GetAndSetConcurrentlyWithFrequentEvictionsAndFIFO(b *testing.B) {
	testValue := strings.Repeat("a", 256)
	cache := NewCache().WithEvictionPolicy(LeastRecentlyUsed).WithMaxSize(3).WithMaxMemoryUsage(NoMaxMemoryUsage)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			k := strconv.Itoa(rand.Intn(15))
			cache.Set(k, testValue)
			_, _ = cache.Get(k)
		}
	})
	b.ReportAllocs()
}

func BenchmarkCache_GetConcurrentlyWithLRU(b *testing.B) {
	testValue := strings.Repeat("a", 256)
	cache := NewCache().WithMaxSize(b.N).WithMaxMemoryUsage(NoMaxMemoryUsage)
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
	b.ReportAllocs()
}

func BenchmarkCache_GetConcurrentlyWithFIFO(b *testing.B) {
	testValue := strings.Repeat("a", 256)
	cache := NewCache().WithMaxSize(b.N).WithMaxMemoryUsage(NoMaxMemoryUsage)
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
	b.ReportAllocs()
}

func BenchmarkCache_GetKeysThatDoNotExistConcurrently(b *testing.B) {
	cache := NewCache().WithMaxSize(1000).WithMaxMemoryUsage(NoMaxMemoryUsage)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, ok := cache.Get(strconv.Itoa(rand.Intn(b.N))); ok {
				b.Errorf("Cache should've been empty")
			}
		}
	})
	b.ReportAllocs()
}
