package gocacheserver

import (
	"fmt"
	"github.com/TwinProduction/gocache"
	"github.com/tidwall/redcon"
	"log"
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

	// Port is the port that the server listens on
	Port int

	AutoSaveInterval time.Duration
	AutoSaveFile     string
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
		defer func() {
			log.Printf("Saving to %s before closing...", server.AutoSaveFile)
			start := time.Now()
			err := server.Cache.SaveToFile(server.AutoSaveFile)
			if err != nil {
				log.Printf("error while autosaving: %s", err.Error())
			}
			log.Printf("Saved successfully in %s", time.Since(start))
		}()
	}
	address := fmt.Sprintf(":%d", DefaultServerPort)
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
			// use this function to accept or deny the connection.
			// log.Printf("accept: %s", conn.RemoteAddr())
			return true
		},
		func(conn redcon.Conn, err error) {
			// this is called when the connection has been closed
			// log.Printf("closed: %s, err: %v", conn.RemoteAddr(), err)
		},
	)
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
	if len(cmd.Args) != 3 {
		conn.WriteError(fmt.Sprintf("ERR wrong number of arguments for '%s' command", string(cmd.Args[0])))
		return
	}
	server.Cache.Set(string(cmd.Args[1]), cmd.Args[2])
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
