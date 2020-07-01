package gocache

import (
	"fmt"
	"testing"
)

func BenchmarkMap_Set(b *testing.B) {
	m := make(map[string]interface{})
	for n := 0; n < b.N; n++ {
		m[fmt.Sprintf("test_%d", n)] = []byte("value")
	}
}

func BenchmarkCache_SetWithMaxSize10(b *testing.B) {
	cache := NewCache().WithMaxSize(10)

	for n := 0; n < b.N; n++ {
		cache.Set(fmt.Sprintf("test_%d", n), []byte("value"))
	}
}

func BenchmarkCache_GetWithMaxSize10(b *testing.B) {
	cache := NewCache().WithMaxSize(10)

	for n := 0; n < b.N; n++ {
		cache.Get(fmt.Sprintf("test_%d", n))
	}
}

func BenchmarkCache_GetAndSetConcurrentlyWithMaxSize10(b *testing.B) {
	data := map[string]string{
		"key_01": "key_01_value",
		"key_02": "key_02_value",
		"key_03": "key_03_value",
		"key_04": "key_04_value",
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
