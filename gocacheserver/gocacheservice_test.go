package gocacheserver

import (
	"github.com/TwinProduction/gocache"
	"github.com/go-redis/redis"
	"testing"
)

var (
	server *Server
	client *redis.Client
)

func init() {
	server = NewServer(gocache.NewCache().WithEvictionPolicy(gocache.LeastRecentlyUsed).WithMaxSize(10000)).WithPort(16162)
	go server.Start()
	client = redis.NewClient(&redis.Options{
		Addr: "localhost:16162",
		DB:   0,
	})
}

func TestParityClientSetCacheGet(t *testing.T) {
	defer server.Cache.Clear()
	const ExpectedValue = "client-set-cache-get"
	client.Set("key", ExpectedValue, 0)
	valueFromCache, ok := server.Cache.Get("key")
	if !ok {
		t.Error("expected key to exist")
	}
	if valueFromCache != ExpectedValue {
		t.Errorf("expected: %s, but got: %s", ExpectedValue, valueFromCache)
	}
}

func TestParityClientSetClientGet(t *testing.T) {
	defer server.Cache.Clear()
	const ExpectedValue = "client-set-client-get"
	client.Set("key", ExpectedValue, 0)
	valueFromRedisClient, err := client.Get("key").Result()
	if err != nil {
		t.Error(err)
	}
	if valueFromRedisClient != ExpectedValue {
		t.Errorf("expected: %s, but got: %s", ExpectedValue, valueFromRedisClient)
	}
}

func TestParityCacheSetClientGet(t *testing.T) {
	defer server.Cache.Clear()
	const ExpectedValue = "cache-set-client-get"
	server.Cache.Set("key", ExpectedValue)
	valueFromRedisClient, err := client.Get("key").Result()
	if err != nil {
		t.Error(err)
	}
	if valueFromRedisClient != ExpectedValue {
		t.Errorf("expected: %s, but got: %s", ExpectedValue, valueFromRedisClient)
	}
}

func TestSET(t *testing.T) {
	defer server.Cache.Clear()
	const ExpectedInitialValue = "v"
	const ExpectedFinalValue = "updated"
	// Set the value for the first time
	client.Set("key", ExpectedInitialValue, 0)
	value, err := client.Get("key").Result()
	if err != nil {
		t.Error(err)
	}
	if value != ExpectedInitialValue {
		t.Errorf("expected: %s, but got: %s", ExpectedInitialValue, value)
	}
	// Update the existing entry
	client.Set("key", ExpectedFinalValue, 0)
	value, err = client.Get("key").Result()
	if err != nil {
		t.Error(err)
	}
	if value != ExpectedFinalValue {
		t.Errorf("expected: %s, but got: %s", ExpectedFinalValue, value)
	}
}

func TestDEL(t *testing.T) {
	defer server.Cache.Clear()
	client.Set("key", "value", 0)
	if _, ok := server.Cache.Get("key"); !ok {
		t.Error("key should've existed")
	}
	client.Del("key")
	if _, ok := server.Cache.Get("key"); ok {
		t.Error("key should've been deleted")
	}
}

func TestMSET(t *testing.T) {
	defer server.Cache.Clear()
	client.MSet("k1", "v1", "k2", "v2")
	if _, ok := server.Cache.Get("k1"); !ok {
		t.Error("k1 should've existed")
	}
	if _, ok := server.Cache.Get("k2"); !ok {
		t.Error("k2 should've existed")
	}
}
