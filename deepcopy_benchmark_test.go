package gocache

import (
	"testing"
)

type BenchTestStruct struct {
	ID      int
	Name    string
	Tags    []string
	Scores  map[string]int
	Details *BenchTestDetails
}

type BenchTestDetails struct {
	Description string
	Values      []int
}

// Benchmarks with deep copy enabled (default)
func BenchmarkWithDeepCopy_Set_SimpleValue(b *testing.B) {
	cache := NewCache().WithDeepCopy(true)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set("key", i)
	}
}

func BenchmarkWithDeepCopy_Get_SimpleValue(b *testing.B) {
	cache := NewCache().WithDeepCopy(true)
	cache.Set("key", 42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("key")
	}
}

func BenchmarkWithDeepCopy_Set_String(b *testing.B) {
	cache := NewCache().WithDeepCopy(true)
	value := "This is a test string with some reasonable length for benchmarking"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set("key", value)
	}
}

func BenchmarkWithDeepCopy_Get_String(b *testing.B) {
	cache := NewCache().WithDeepCopy(true)
	cache.Set("key", "This is a test string with some reasonable length for benchmarking")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("key")
	}
}

func BenchmarkWithDeepCopy_Set_ComplexStruct(b *testing.B) {
	cache := NewCache().WithDeepCopy(true)
	value := BenchTestStruct{
		ID:     123,
		Name:   "Test Item",
		Tags:   []string{"tag1", "tag2", "tag3"},
		Scores: map[string]int{"math": 90, "science": 85, "english": 88},
		Details: &BenchTestDetails{
			Description: "This is a detailed description",
			Values:      []int{1, 2, 3, 4, 5},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set("key", value)
	}
}

func BenchmarkWithDeepCopy_Get_ComplexStruct(b *testing.B) {
	cache := NewCache().WithDeepCopy(true)
	cache.Set("key", BenchTestStruct{
		ID:     123,
		Name:   "Test Item",
		Tags:   []string{"tag1", "tag2", "tag3"},
		Scores: map[string]int{"math": 90, "science": 85, "english": 88},
		Details: &BenchTestDetails{
			Description: "This is a detailed description",
			Values:      []int{1, 2, 3, 4, 5},
		},
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("key")
	}
}

// Benchmarks with deep copy disabled
func BenchmarkWithoutDeepCopy_Set_SimpleValue(b *testing.B) {
	cache := NewCache().WithDeepCopy(false)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set("key", i)
	}
}

func BenchmarkWithoutDeepCopy_Get_SimpleValue(b *testing.B) {
	cache := NewCache().WithDeepCopy(false)
	cache.Set("key", 42)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("key")
	}
}

func BenchmarkWithoutDeepCopy_Set_String(b *testing.B) {
	cache := NewCache().WithDeepCopy(false)
	value := "This is a test string with some reasonable length for benchmarking"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set("key", value)
	}
}

func BenchmarkWithoutDeepCopy_Get_String(b *testing.B) {
	cache := NewCache().WithDeepCopy(false)
	cache.Set("key", "This is a test string with some reasonable length for benchmarking")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("key")
	}
}

func BenchmarkWithoutDeepCopy_Set_ComplexStruct(b *testing.B) {
	cache := NewCache().WithDeepCopy(false)
	value := BenchTestStruct{
		ID:     123,
		Name:   "Test Item",
		Tags:   []string{"tag1", "tag2", "tag3"},
		Scores: map[string]int{"math": 90, "science": 85, "english": 88},
		Details: &BenchTestDetails{
			Description: "This is a detailed description",
			Values:      []int{1, 2, 3, 4, 5},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set("key", value)
	}
}

func BenchmarkWithoutDeepCopy_Get_ComplexStruct(b *testing.B) {
	cache := NewCache().WithDeepCopy(false)
	cache.Set("key", BenchTestStruct{
		ID:     123,
		Name:   "Test Item",
		Tags:   []string{"tag1", "tag2", "tag3"},
		Scores: map[string]int{"math": 90, "science": 85, "english": 88},
		Details: &BenchTestDetails{
			Description: "This is a detailed description",
			Values:      []int{1, 2, 3, 4, 5},
		},
	})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get("key")
	}
}
