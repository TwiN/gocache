package gocacheserver

import (
	"bytes"
	"fmt"
	"github.com/TwinProduction/gocache"
	"github.com/tidwall/redcon"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultServerPort is the default port for the server
	DefaultServerPort = 6379
)

// Server is a cache server using gocache as cache and RESP (Redis bindings) as server
type Server struct {
	Cache *gocache.Cache

	Port int

	AutoSaveInterval time.Duration
	AutoSaveFile     string

	startTime           time.Time
	numberOfConnections int
}

// NewServer creates a new cache server
func NewServer(cache *gocache.Cache) *Server {
	return &Server{
		Cache: cache,
		Port:  DefaultServerPort,
	}
}

// WithAutoSave allows the configuration of the autosave feature interval at which
// the cache will be automatically saved
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
func (server *Server) Start() error {
	if server.AutoSaveInterval != 0 {
		go server.autoSave()
	}
	if err := server.Cache.StartJanitor(); err != nil {
		panic(err)
	}
	address := fmt.Sprintf(":%d", server.Port)
	server.startTime = time.Now()
	log.Printf("Listening on %s", address)
	err := redcon.ListenAndServe(address,
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
	server.Cache.StopJanitor()
	if server.AutoSaveInterval != 0 {
		log.Printf("Saving to %s before closing...", server.AutoSaveFile)
		start := time.Now()
		err := server.Cache.SaveToFile(server.AutoSaveFile)
		if err != nil {
			log.Printf("error while autosaving: %s", err.Error())
		}
		log.Printf("Saved successfully in %s", time.Since(start))
	}
	return err
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
		server.Cache.Set(string(cmd.Args[1]), cmd.Args[2])
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
	server.Cache.SetWithTTL(string(cmd.Args[1]), cmd.Args[3], time.Duration(unit)*time.Second)
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
	keyValues := server.Cache.GetAll(keys)
	if len(keyValues) != len(keys) {
		conn.WriteError(fmt.Sprintf("ERR internal error, expected %d keys, got %d instead", len(keys), len(keyValues)))
	}
	conn.WriteArray(len(keyValues))
	for _, key := range keys {
		conn.WriteAny(keyValues[key])
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
		buffer.WriteString(fmt.Sprintf("evicted_keys:%d\n", server.Cache.Stats.EvictedKeys))
		buffer.WriteString(fmt.Sprintf("expired_keys:%d\n", server.Cache.Stats.ExpiredKeys))
		buffer.WriteString(fmt.Sprintf("current_keys:%d\n", server.Cache.Count()))
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

// autoSave automatically saves every AutoSaveInterval
func (server *Server) autoSave() {
	for {
		time.Sleep(server.AutoSaveInterval)
		err := server.Cache.SaveToFile(server.AutoSaveFile)
		if err != nil {
			log.Printf("error while autosaving: %s", err.Error())
			continue
		}
	}
}
