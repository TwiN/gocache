package gocache

import (
	"fmt"
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
	b.ReportAllocs()
}

func BenchmarkMap_Set(b *testing.B) {
	values := map[string]string{
		"small":  "a",
		"medium": strings.Repeat("a", 1024),
		"large":  strings.Repeat("a", 1024*100),
	}
	for name, value := range values {
		b.Run(fmt.Sprintf("%s value", name), func(b *testing.B) {
			m := make(map[string]interface{})
			for n := 0; n < b.N; n++ {
				m[strconv.Itoa(n)] = value
			}
			b.ReportAllocs()
		})
	}
}

func BenchmarkCache_Get(b *testing.B) {
	evictionPolicies := []EvictionPolicy{FirstInFirstOut, LeastRecentlyUsed}
	for _, evictionPolicy := range evictionPolicies {
		cache := NewCache().WithMaxSize(NoMaxSize).WithMaxMemoryUsage(NoMaxMemoryUsage)
		b.Run(string(evictionPolicy), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				cache.Get(strconv.Itoa(n))
			}
			b.ReportAllocs()
		})
	}
}

func BenchmarkCache_Set(b *testing.B) {
	values := map[string]string{
		"small":  "a",
		"medium": strings.Repeat("a", 1024),
		"large":  strings.Repeat("a", 1024*100),
	}
	evictionPolicies := []EvictionPolicy{FirstInFirstOut, LeastRecentlyUsed}
	for _, evictionPolicy := range evictionPolicies {
		for name, value := range values {
			b.Run(fmt.Sprintf("%s %s value", evictionPolicy, name), func(b *testing.B) {
				cache := NewCache().WithMaxSize(NoMaxSize).WithMaxMemoryUsage(NoMaxMemoryUsage).WithEvictionPolicy(evictionPolicy)
				for n := 0; n < b.N; n++ {
					cache.Set(strconv.Itoa(n), value)
				}
				b.ReportAllocs()
			})
		}
	}
}

// BenchmarkCache_SetUsingMaxMemoryUsage does NOT test evictions, it tests the overhead of the extra work
// automatically performed when using MaxMemoryUsage
func BenchmarkCache_SetUsingMaxMemoryUsage(b *testing.B) {
	values := map[string]string{
		"small":  "a",
		"medium": strings.Repeat("a", 1024),
		"large":  strings.Repeat("a", 1024*100),
	}
	for name, value := range values {
		b.Run(fmt.Sprintf("%s value", name), func(b *testing.B) {
			cache := NewCache().WithMaxSize(NoMaxSize).WithMaxMemoryUsage(999 * Gigabyte)
			for n := 0; n < b.N; n++ {
				cache.Set(strconv.Itoa(n), value)
			}
			b.ReportAllocs()
		})
	}
}

func BenchmarkCache_SetWithMaxSize(b *testing.B) {
	values := map[string]string{
		"small":  "a",
		"medium": strings.Repeat("a", 1024),
		"large":  strings.Repeat("a", 1024*100),
	}
	maxSizes := []int{100, 10000, 100000}
	for name, value := range values {
		for _, maxSize := range maxSizes {
			b.Run(fmt.Sprintf("%d %s value", maxSize, name), func(b *testing.B) {
				cache := NewCache().WithMaxSize(maxSize)
				for n := 0; n < b.N; n++ {
					cache.Set(strconv.Itoa(n), value)
				}
				b.ReportAllocs()
			})
		}
	}
}

func BenchmarkCache_SetWithMaxSizeAndLRU(b *testing.B) {
	values := map[string]string{
		"small":  "a",
		"medium": strings.Repeat("a", 1024),
		"large":  strings.Repeat("a", 1024*100),
	}
	maxSizes := []int{100, 10000, 100000}
	for name, value := range values {
		for _, maxSize := range maxSizes {
			b.Run(fmt.Sprintf("%d %s value", maxSize, name), func(b *testing.B) {
				cache := NewCache().WithMaxSize(maxSize).WithEvictionPolicy(LeastRecentlyUsed)
				for n := 0; n < b.N; n++ {
					cache.Set(strconv.Itoa(n), value)
				}
				b.ReportAllocs()
			})
		}
	}
}

func BenchmarkCache_GetSetMultipleConcurrent(b *testing.B) {
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

func BenchmarkCache_GetSetConcurrentWithFrequentEviction(b *testing.B) {
	value := strings.Repeat("a", 256)
	evictionPolicies := []EvictionPolicy{FirstInFirstOut, LeastRecentlyUsed}
	for _, evictionPolicy := range evictionPolicies {
		b.Run(string(evictionPolicy), func(b *testing.B) {
			cache := NewCache().WithEvictionPolicy(LeastRecentlyUsed).WithMaxSize(3).WithMaxMemoryUsage(NoMaxMemoryUsage)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					k := strconv.Itoa(rand.Intn(15))
					cache.Set(k, value)
					_, _ = cache.Get(k)
				}
			})
			b.ReportAllocs()
		})

	}
}

func BenchmarkCache_GetConcurrentWithLRU(b *testing.B) {
	value := strings.Repeat("a", 256)
	for _, evictionPolicy := range []EvictionPolicy{FirstInFirstOut, LeastRecentlyUsed} {
		b.Run(string(evictionPolicy), func(b *testing.B) {
			cache := NewCache().WithMaxSize(100000)
			for i := 0; i < 100000; i++ {
				cache.Set(strconv.Itoa(i), value)
			}
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					key := strconv.Itoa(rand.Intn(100000))
					val, ok := cache.Get(key)
					if !ok {
						b.Errorf("key: %v; value: %v", key, val)
					}
					if val != value {
						b.Errorf("expected: %v; got: %v", val, value)
					}
				}
			})
			b.ReportAllocs()
		})
	}
}

// Note: The default value for Cache.forceNilInterfaceOnNilPointer is true
func BenchmarkCache_WithForceNilInterfaceOnNilPointer(b *testing.B) {
	const (
		Min = 10000
		Max = 99999
	)
	type Struct struct {
		Value string
	}
	forceNilInterfaceOnNilPointerValues := []bool{true, false}
	values := []*Struct{nil, {Value: "value"}}
	for _, forceNilInterfaceOnNilPointer := range forceNilInterfaceOnNilPointerValues {
		for _, value := range values {
			name := fmt.Sprintf("%v", forceNilInterfaceOnNilPointer)
			if value == nil {
				name += " with nil struct pointer"
			}
			b.Run(name, func(b *testing.B) {
				cache := NewCache().WithMaxSize(NoMaxSize).WithMaxMemoryUsage(NoMaxMemoryUsage).WithForceNilInterfaceOnNilPointer(forceNilInterfaceOnNilPointer)
				for n := 0; n < b.N; n++ {
					cache.Set(strconv.Itoa(rand.Intn(Max-Min)+Min), value)
				}
				b.ReportAllocs()
			})
		}
	}
}

func BenchmarkCache_WithForceNilInterfaceOnNilPointerWithConcurrency(b *testing.B) {
	const (
		Min = 10000
		Max = 99999
	)
	type Struct struct {
		Value string
	}
	forceNilInterfaceOnNilPointerValues := []bool{true, false}
	values := []*Struct{nil, {Value: "value"}}
	for _, forceNilInterfaceOnNilPointer := range forceNilInterfaceOnNilPointerValues {
		for _, value := range values {
			name := fmt.Sprintf("%v", forceNilInterfaceOnNilPointer)
			if value == nil {
				name += " with nil struct pointer"
			}
			b.Run(name, func(b *testing.B) {
				cache := NewCache().WithMaxSize(NoMaxSize).WithMaxMemoryUsage(NoMaxMemoryUsage).WithForceNilInterfaceOnNilPointer(forceNilInterfaceOnNilPointer)
				b.RunParallel(func(pb *testing.PB) {
					for pb.Next() {
						cache.Set(strconv.Itoa(rand.Intn(Max-Min)+Min), value)
					}
				})
				b.ReportAllocs()
			})
		}
	}
}
