// +build !race

package gocacheserver

import (
	"github.com/TwinProduction/gocache"
	"github.com/go-redis/redis"
	"os"
	"testing"
	"time"
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

func TestSET_PX(t *testing.T) {
	defer server.Cache.Clear()
	const ExpectedValue = "v"
	client.Set("key", ExpectedValue, 9999*time.Millisecond)
	value, err := client.Get("key").Result()
	if err != nil {
		t.Error(err)
	}
	if value != ExpectedValue {
		t.Errorf("expected: %s, but got: %s", ExpectedValue, value)
	}
	ttl, _ := server.Cache.TTL("key")
	if ttl.Seconds() < 9 || ttl.Seconds() > 10 {
		t.Error("expected TTL of ~9999ms")
	}
}

func TestSET_EX(t *testing.T) {
	defer server.Cache.Clear()
	const ExpectedValue = "v"
	client.Set("key", ExpectedValue, 10*time.Second)
	value, err := client.Get("key").Result()
	if err != nil {
		t.Error(err)
	}
	if value != ExpectedValue {
		t.Errorf("expected: %s, but got: %s", ExpectedValue, value)
	}
	ttl, _ := server.Cache.TTL("key")
	if ttl.Seconds() < 8 || ttl.Seconds() > 10 {
		t.Error("expected TTL of ~10s")
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

func TestEXPIRE(t *testing.T) {
	defer server.Cache.Clear()
	client.Set("key", "value", 0)
	if _, ok := server.Cache.Get("key"); !ok {
		t.Error("key should've existed")
	}
	// expire the key now
	client.Expire("key", 0)
	// wait a bit to make sure the key's gone
	time.Sleep(time.Millisecond)
	if _, ok := server.Cache.Get("key"); ok {
		t.Error("key should've expired")
	}
}

func TestSETEX(t *testing.T) {
	defer server.Cache.Clear()
	// SETEX doesn't exist in the library, see https://github.com/go-redis/redis/pull/1546
	client.Do("SETEX", "key", time.Hour.Seconds(), "value").Val()
	if _, ok := server.Cache.Get("key"); !ok {
		t.Error("key should've existed")
	}
	ttl, _ := server.Cache.TTL("key")
	if ttl.Minutes() < 59 || ttl.Minutes() > 60 {
		t.Error("key should've had a TTL between 59 and 60 minutes")
	}
}

func TestEXISTS(t *testing.T) {
	defer server.Cache.Clear()
	client.Set("k1", "v1", 0)
	client.Set("k2", "v2", 0)
	client.Set("k3", "v3", 0)
	output := client.Exists("k1", "k2", "key-that-does-not-exist").Val()
	if output != 2 {
		t.Error("Expected 2 keys to exist, got", output)
	}
}

func TestFLUSHDB(t *testing.T) {
	defer server.Cache.Clear()
	server.Cache.Set("key", "value")
	if server.Cache.Count() != 1 {
		t.Error("cache should have a size of 1")
	}
	client.FlushDB()
	if server.Cache.Count() != 0 {
		t.Error("cache should've been cleared")
	}
}

func TestPING(t *testing.T) {
	if client.Ping().Val() != "PONG" {
		t.Error("Server should've been able to pong :(")
	}
}

func TestECHO(t *testing.T) {
	if client.Echo("hey").Val() != "hey" {
		t.Error("Server should've been able to echo")
	}
}

func TestINFO(t *testing.T) {
	if len(client.Info().Val()) < 100 {
		t.Error("INFO should've returned at least some info")
	}
}

func TestSCAN(t *testing.T) {
	defer server.Cache.Clear()
	server.Cache.Set("vegetable", "true")
	server.Cache.Set("k1", "value")
	server.Cache.Set("k2", "value")
	server.Cache.Set("fruit", "true")
	if server.Cache.Count() != 4 {
		t.Error("cache should have a size of 4")
	}
	keys, cursor := client.Scan(0, "k*", 9999).Val()
	if cursor != 0 {
		t.Error("cursor returned should've been 0, because it isn't supported yet")
	}
	if len(keys) != 2 {
		t.Error("should've returned 2 keys")
	}
	for _, k := range keys {
		if k != "k1" && k != "k2" {
			t.Error("key should've been k1 or k2, but was", k)
		}
	}
}

func TestSCAN_AndRespectCount(t *testing.T) {
	defer server.Cache.Clear()
	server.Cache.Set("vegetable", "true")
	server.Cache.Set("k1", "value")
	server.Cache.Set("k2", "value")
	server.Cache.Set("fruit", "true")
	if server.Cache.Count() != 4 {
		t.Error("cache should have a size of 4")
	}
	keys, cursor := client.Scan(0, "k*", 1).Val()
	if cursor != 0 {
		t.Error("cursor returned should've been 0, because it isn't supported yet")
	}
	if len(keys) != 1 {
		t.Error("should've returned 1 key, because the limit was set to 1")
	}
}

func TestTTL(t *testing.T) {
	defer server.Cache.Clear()
	client.Set("key", "value", 10*time.Second)
	ttl := client.TTL("key").Val()
	if ttl.Seconds() < 9 || ttl.Seconds() > 10 {
		t.Error("expected TTL of ~9999ms")
	}
}

func TestServer_WithAutoSave(t *testing.T) {
	defer os.Remove("TestServer_WithAutoSave.bak")
	serverWithAutoSave := NewServer(gocache.NewCache().WithEvictionPolicy(gocache.LeastRecentlyUsed).WithMaxSize(10)).WithPort(16163).WithAutoSave(10*time.Millisecond, "TestServer_WithAutoSave.bak")
	go serverWithAutoSave.Start()
	serverWithAutoSave.Cache.Set("john", "doe")
	serverWithAutoSave.Cache.Set("jane", "doe")
	// Wait long enough for the auto save to be triggered
	time.Sleep(30 * time.Millisecond)
	// Stop the server
	serverWithAutoSave.Stop()
	for {
		if !serverWithAutoSave.running {
			break
		}
		time.Sleep(time.Millisecond)
	}
	// We'll start another server with the save configuration as the first server.
	// This should trigger the data from the first server to be retrieved from the AutoSaveFile into the new server.
	otherServerWithAutoSave := NewServer(gocache.NewCache().WithEvictionPolicy(gocache.LeastRecentlyUsed).WithMaxSize(10)).WithPort(16163).WithAutoSave(10*time.Minute, "TestServer_WithAutoSave.bak")
	go otherServerWithAutoSave.Start()
	// Wait for long enough to the cache to be re-populated
	for {
		if otherServerWithAutoSave.running {
			break
		}
		time.Sleep(time.Millisecond)
	}
	if otherServerWithAutoSave.Cache.Count() != 2 {
		t.Errorf("New cache server should've been repopulated by the AutoSaveFile of and have a size of 2, but has %d instead", otherServerWithAutoSave.Cache.Count())
	}
}
