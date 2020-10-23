package gocacheserver

import (
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func BenchmarkSETEX(b *testing.B) {
	defer server.Cache.Clear()
	for n := 0; n < b.N; n++ {
		// SETEX doesn't exist in the library, see https://github.com/go-redis/redis/pull/1546
		client.Do("SETEX", strconv.Itoa(n), time.Hour.Seconds(), "value").Val()
	}
}

func BenchmarkSET(b *testing.B) {
	defer server.Cache.Clear()
	for n := 0; n < b.N; n++ {
		client.Set(strconv.Itoa(n), "value", time.Hour).Val()
	}
}

func BenchmarkSETGET(b *testing.B) {
	defer server.Cache.Clear()
	for n := 0; n < b.N; n++ {
		client.Set(strconv.Itoa(n), "value", time.Hour).Val()
		client.Get(strconv.Itoa(n)).Val()
	}
}

func BenchmarkSETConcurrently(b *testing.B) {
	defer server.Cache.Clear()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key := strconv.Itoa(rand.Intn(b.N))
			client.Set(key, "value", time.Hour).Val()
		}
	})
}

func BenchmarkSETGETConcurrently(b *testing.B) {
	defer server.Cache.Clear()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key := strconv.Itoa(rand.Intn(b.N))
			client.Set(key, "value", 0).Val()
			client.Get(key).Val()
		}
	})
}
