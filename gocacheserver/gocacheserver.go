package gocacheserver

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/TwinProduction/gocache"
	"github.com/tidwall/redcon"
)

const (
	// DefaultServerPort is the default port for the server
	DefaultServerPort = 6379
)

// Server is a cache server using gocache as cache and RESP (Redis bindings) as server
type Server struct {
	// Cache is the actual cache
	Cache *gocache.Cache

	// Port is the port that the server will listen on
	Port int

	// AutoSaveInterval is the interval at which the server will automatically save the Cache
	AutoSaveInterval time.Duration

	// AutoSaveFile is the file in which the cache will be persisted every AutoSaveInterval
	AutoSaveFile string

	startTime           time.Time
	numberOfConnections int

	running     bool
	cacheServer *redcon.Server
}

// NewServer creates a new cache server
func NewServer(cache *gocache.Cache) *Server {
	return &Server{
		Cache: cache,
		Port:  DefaultServerPort,
	}
}

// WithAutoSave allows the configuration of the automatic saving feature.
// Note that setting this will also cause the server to immediately read the file passed and populate the cache
//
// Disabled if set to 0
func (server *Server) WithAutoSave(interval time.Duration, file string) *Server {
	server.AutoSaveInterval = interval
	server.AutoSaveFile = file
	return server
}

// WithPort sets the port of the server
func (server *Server) WithPort(port int) *Server {
	server.Port = port
	return server
}

// Start starts the cache server, which includes the autosave
//
// This is a blocking function, therefore, you are expected to run this on a goroutine
func (server *Server) Start() error {
	if server.AutoSaveInterval != 0 {
		err := server.loadAutoSaveFileIfExists()
		if err != nil {
			return fmt.Errorf("ran into the following error while attempting to load the auto save file in memory: %s", err.Error())
		}
		go server.autoSave()
	}
	if err := server.Cache.StartJanitor(); err != nil {
		return err
	}
	address := fmt.Sprintf(":%d", server.Port)
	server.cacheServer = redcon.NewServer(address,
		func(conn redcon.Conn, cmd redcon.Command) {
			switch strings.ToUpper(string(cmd.Args[0])) {
			case "GET":
				server.get(cmd, conn)
			case "SET":
				server.set(cmd, conn)
			case "DEL":
				server.del(cmd, conn)
			case "EXISTS":
				server.exists(cmd, conn)
			case "MGET":
				server.mget(cmd, conn)
			case "MSET":
				server.mset(cmd, conn)
			case "SCAN":
				server.scan(cmd, conn)
			case "TTL":
				server.ttl(cmd, conn)
			case "EXPIRE":
				server.expire(cmd, conn)
			case "SETEX":
				server.setex(cmd, conn)
			case "FLUSHDB":
				server.flushDb(cmd, conn)
			case "INFO":
				server.info(cmd, conn)
			case "PING":
				conn.WriteString("PONG")
			case "QUIT":
				conn.WriteString("OK")
				conn.Close()
			case "ECHO":
				if len(cmd.Args) != 2 {
					conn.WriteError(fmt.Sprintf("ERR wrong number of arguments for '%s' command", string(cmd.Args[0])))
					return
				}
				conn.WriteBulk(cmd.Args[1])
			default:
				conn.WriteError(fmt.Sprintf("ERR unknown command '%s'", string(cmd.Args[0])))
			}
		},
		func(conn redcon.Conn) bool {
			server.numberOfConnections += 1
			return true
		},
		func(conn redcon.Conn, err error) {
			server.numberOfConnections -= 1
		},
	)
	server.startTime = time.Now()
	server.running = true
	log.Printf("Listening on %s", address)
	err := server.cacheServer.ListenAndServe()
	server.Cache.StopJanitor()
	server.running = false
	if server.AutoSaveInterval != 0 {
		log.Printf("Saving to %s before closing...", server.AutoSaveFile)
		start := time.Now()
		if err := server.Cache.SaveToFile(server.AutoSaveFile); err != nil {
			log.Printf("error while autosaving: %s", err.Error())
		}
		log.Printf("Saved successfully in %s", time.Since(start))
	}
	return err
}

// Stop closes the Server
func (server *Server) Stop() error {
	if server.cacheServer == nil {
		// If the cache server is nil, there's nothing to stop.
		return nil
	}
	return server.cacheServer.Close()
}

func (server *Server) get(cmd redcon.Command, conn redcon.Conn) {
	if len(cmd.Args) != 2 {
		conn.WriteError(fmt.Sprintf("ERR wrong number of arguments for '%s' command", string(cmd.Args[0])))
		return
	}
	val, ok := server.Cache.Get(string(cmd.Args[1]))
	if !ok {
		conn.WriteNull()
	} else {
		conn.WriteAny(val)
	}
}

func (server *Server) set(cmd redcon.Command, conn redcon.Conn) {
	numberOfArguments := len(cmd.Args)
	if numberOfArguments != 3 && numberOfArguments != 5 && numberOfArguments != 6 {
		conn.WriteError(fmt.Sprintf("ERR wrong number of arguments for '%s' command", string(cmd.Args[0])))
		return
	}
	if numberOfArguments == 3 {
		server.Cache.Set(string(cmd.Args[1]), string(cmd.Args[2]))
	} else {
		unit, err := strconv.Atoi(string(cmd.Args[4]))
		if err != nil {
			conn.WriteError("ERR value is not an integer or out of range")
			return
		}
		option := strings.ToUpper(string(cmd.Args[3]))
		if option == "EX" {
			server.Cache.SetWithTTL(string(cmd.Args[1]), cmd.Args[2], time.Duration(unit)*time.Second)
		} else if option == "PX" {
			server.Cache.SetWithTTL(string(cmd.Args[1]), cmd.Args[2], time.Duration(unit)*time.Millisecond)
		} else {
			conn.WriteError("ERR syntax error")
			return
		}
	}
	conn.WriteString("OK")
}

func (server *Server) setex(cmd redcon.Command, conn redcon.Conn) {
	if len(cmd.Args) != 4 {
		conn.WriteError(fmt.Sprintf("ERR wrong number of arguments for '%s' command", string(cmd.Args[0])))
		return
	}
	unit, err := strconv.Atoi(string(cmd.Args[2]))
	if err != nil {
		conn.WriteError("ERR value is not an integer or out of range")
		return
	}
	server.Cache.SetWithTTL(string(cmd.Args[1]), string(cmd.Args[3]), time.Duration(unit)*time.Second)
	conn.WriteString("OK")
}

func (server *Server) del(cmd redcon.Command, conn redcon.Conn) {
	if len(cmd.Args) < 2 {
		conn.WriteError(fmt.Sprintf("ERR wrong number of arguments for '%s' command", string(cmd.Args[0])))
		return
	}
	numberOfKeysDeleted := 0
	for index := range cmd.Args {
		if index == 0 {
			continue
		}
		ok := server.Cache.Delete(string(cmd.Args[index]))
		if ok {
			numberOfKeysDeleted++
		}
	}
	conn.WriteInt(numberOfKeysDeleted)
}

func (server *Server) exists(cmd redcon.Command, conn redcon.Conn) {
	if len(cmd.Args) < 2 {
		conn.WriteError(fmt.Sprintf("ERR wrong number of arguments for '%s' command", string(cmd.Args[0])))
		return
	}
	numberOfExistingKeys := 0
	for index := range cmd.Args {
		if index == 0 {
			continue
		}
		_, ok := server.Cache.Get(string(cmd.Args[index]))
		if ok {
			numberOfExistingKeys++
		}
	}
	conn.WriteInt(numberOfExistingKeys)
}

func (server *Server) mget(cmd redcon.Command, conn redcon.Conn) {
	if len(cmd.Args) < 2 {
		conn.WriteError(fmt.Sprintf("ERR wrong number of arguments for '%s' command", string(cmd.Args[0])))
		return
	}
	var keys []string
	for index := range cmd.Args {
		if index == 0 {
			continue
		}
		keys = append(keys, string(cmd.Args[index]))
	}
	keyValues := server.Cache.GetByKeys(keys)
	if len(keyValues) != len(keys) {
		conn.WriteError(fmt.Sprintf("ERR internal error, expected %d keys, got %d instead", len(keys), len(keyValues)))
	}
	conn.WriteArray(len(keyValues))
	for _, key := range keys {
		conn.WriteAny(keyValues[key])
	}
}

func (server *Server) mset(cmd redcon.Command, conn redcon.Conn) {
	if len(cmd.Args) < 3 {
		conn.WriteError(fmt.Sprintf("ERR wrong number of arguments for '%s' command", string(cmd.Args[0])))
		return
	}
	newEntries := make(map[string]interface{})
	for index := range cmd.Args {
		if index == 0 {
			continue
		}
		if index%2 == 0 {
			key := string(cmd.Args[index-1])
			value := string(cmd.Args[index])
			newEntries[key] = value
		}
	}
	server.Cache.SetAll(newEntries)
	conn.WriteString("OK")
}

// scan is used to search keys by pattern
// At the moment, the cursor is ignored.
func (server *Server) scan(cmd redcon.Command, conn redcon.Conn) {
	numberOfArguments := len(cmd.Args)
	if numberOfArguments != 2 && numberOfArguments != 4 && numberOfArguments != 6 {
		conn.WriteError(fmt.Sprintf("ERR wrong number of arguments for '%s' command", string(cmd.Args[0])))
		return
	}
	// XXX: The cursor is currently ignored, but we'll still validate it
	_, err := strconv.Atoi(string(cmd.Args[1]))
	if err != nil {
		conn.WriteError("ERR value is not an integer or out of range")
		return
	}
	var keys []string
	if numberOfArguments == 2 {
		keys = server.Cache.GetKeysByPattern("*", 10)
	} else {
		var (
			count              = 10
			pattern            = "*"
			isConfiguringCount = false
			isConfiguringMatch = false
		)
		for index := range cmd.Args {
			if index < 2 {
				continue
			}
			switch strings.ToUpper(string(cmd.Args[index])) {
			case "MATCH":
				isConfiguringCount = false
				isConfiguringMatch = true
			case "COUNT":
				isConfiguringCount = true
				isConfiguringMatch = false
			default:
				if isConfiguringCount {
					isConfiguringCount = false
					count, err = strconv.Atoi(string(cmd.Args[index]))
					if err != nil {
						conn.WriteError("ERR value is not an integer or out of range")
						return
					}
				} else if isConfiguringMatch {
					isConfiguringMatch = false
					pattern = string(cmd.Args[index])
				} else {
					conn.WriteError("ERR syntax error")
					return
				}
			}
		}
		keys = server.Cache.GetKeysByPattern(pattern, count)
	}
	conn.WriteArray(2)
	// The first value is the cursor used in the previous call. Since we don't support cursors at the moment, we'll
	// hardcode this to 0.
	// This is to prevent automated libraries from looping forever:
	//     An iteration starts when the cursor is set to 0, and terminates when the cursor returned by the server is 0.
	//                                                                        reference: https://redis.io/commands/scan
	conn.WriteAny(0)
	conn.WriteArray(len(keys))
	for _, key := range keys {
		conn.WriteAny(key)
	}
}

func (server *Server) ttl(cmd redcon.Command, conn redcon.Conn) {
	if len(cmd.Args) != 2 {
		conn.WriteError(fmt.Sprintf("ERR wrong number of arguments for '%s' command", string(cmd.Args[0])))
		return
	}
	ttl, err := server.Cache.TTL(string(cmd.Args[1]))
	if err != nil {
		if err == gocache.ErrKeyDoesNotExist {
			conn.WriteInt(-2)
		} else if err == gocache.ErrKeyHasNoExpiration {
			conn.WriteInt(-1)
		} else {
			conn.WriteError(fmt.Sprintf("ERR %s", err.Error()))
		}
		return
	}
	conn.WriteInt(int(ttl.Seconds()))
}

func (server *Server) expire(cmd redcon.Command, conn redcon.Conn) {
	if len(cmd.Args) != 3 {
		conn.WriteError(fmt.Sprintf("ERR wrong number of arguments for '%s' command", string(cmd.Args[0])))
		return
	}
	key := string(cmd.Args[1])
	seconds, err := strconv.Atoi(string(cmd.Args[2]))
	if err != nil {
		conn.WriteError("ERR value is not an integer or out of range")
		return
	}
	updatedSuccessfully := server.Cache.Expire(key, time.Second*time.Duration(seconds))
	if updatedSuccessfully {
		conn.WriteInt(1)
	} else {
		conn.WriteInt(0)
	}
}

func (server *Server) info(cmd redcon.Command, conn redcon.Conn) {
	if len(cmd.Args) > 2 {
		conn.WriteError(fmt.Sprintf("ERR wrong number of arguments for '%s' command", string(cmd.Args[0])))
		return
	}
	var section string
	if len(cmd.Args) == 1 {
		section = "ALL"
	} else {
		section = strings.ToUpper(string(cmd.Args[1]))
	}
	buffer := new(bytes.Buffer)
	if section == "ALL" || section == "SERVER" {
		buffer.WriteString("# Server\n")
		buffer.WriteString(fmt.Sprintf("process_id:%d\n", os.Getpid()))
		buffer.WriteString(fmt.Sprintf("uptime_in_seconds:%d\n", int64(time.Since(server.startTime).Seconds())))
		buffer.WriteString(fmt.Sprintf("uptime_in_days:%d\n", int64(time.Since(server.startTime).Hours()/24)))
		buffer.WriteString("\n")
	}
	if section == "ALL" || section == "CLIENTS" {
		buffer.WriteString("# Clients\n")
		buffer.WriteString(fmt.Sprintf("connected_clients:%d\n", server.numberOfConnections))
		buffer.WriteString("\n")
	}
	if section == "ALL" || section == "STATS" {
		buffer.WriteString("# Stats\n")
		buffer.WriteString(fmt.Sprintf("current_keys:%d\n", server.Cache.Count()))
		buffer.WriteString(fmt.Sprintf("evicted_keys:%d\n", server.Cache.Stats().EvictedKeys))
		buffer.WriteString(fmt.Sprintf("expired_keys:%d\n", server.Cache.Stats().ExpiredKeys))
		buffer.WriteString(fmt.Sprintf("keyspace_hits:%d\n", server.Cache.Stats().Hits))
		buffer.WriteString(fmt.Sprintf("keyspace_misses:%d\n", server.Cache.Stats().Misses))
		buffer.WriteString("\n")
	}
	if section == "ALL" || section == "MEMORY" {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		buffer.WriteString("# Memory\n")
		buffer.WriteString(fmt.Sprintf("used_memory:%d\n", m.HeapSys))
		buffer.WriteString(fmt.Sprintf("used_memory_human:%dM\n", m.HeapSys/1024/1024))
		buffer.WriteString(fmt.Sprintf("used_memory_dataset:%d\n", server.Cache.MemoryUsage()))
		buffer.WriteString(fmt.Sprintf("used_memory_dataset_human:%dM\n", server.Cache.MemoryUsage()/1024/1024))
		buffer.WriteString("\n")
	}
	if section == "ALL" || section == "REPLICATION" {
		buffer.WriteString("# Replication\n")
		buffer.WriteString("role:master\n")
		buffer.WriteString("\n")
	}
	conn.WriteBulkString(fmt.Sprintf("%s\n", strings.TrimSpace(buffer.String())))
}

func (server *Server) flushDb(_ redcon.Command, conn redcon.Conn) {
	server.Cache.Clear()
	conn.WriteString("OK")
}

// loadAutoSaveFileIfExists loads the Cache with the entries present in the AutoSaveFile
func (server *Server) loadAutoSaveFileIfExists() error {
	numberOfEntriesEvicted, err := server.Cache.ReadFromFile(server.AutoSaveFile)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("Server will start with an empty cache, because the specified auto save file doesn't exist")
		} else {
			return err
		}
	}
	if numberOfEntriesEvicted > 0 {
		log.Printf("%d keys had to be evicted after reading the file in order to respect the maximum cache size", numberOfEntriesEvicted)
	}
	if cacheSize := server.Cache.Count(); cacheSize > 0 {
		log.Printf("%d keys loaded into memory from auto save file '%s'", cacheSize, server.AutoSaveFile)
	}
	return nil
}

// autoSave persists the cache to AutoSaveFile every AutoSaveInterval
func (server *Server) autoSave() {
	for {
		time.Sleep(server.AutoSaveInterval)
		if !server.running {
			log.Println("terminating auto save process because server is no longer running")
			break
		}
		start := time.Now()
		log.Printf("Persisting data to %s...", server.AutoSaveFile)
		err := server.Cache.SaveToFile(server.AutoSaveFile)
		if err != nil {
			log.Printf("error while autosaving: %s", err.Error())
			continue
		}
		log.Printf("Persisted data to %s successfully in %s", server.AutoSaveFile, time.Since(start))
	}
}
