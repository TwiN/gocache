package gocache

import "time"

type Entry struct {
	Key   string
	Value interface{}

	// RelevantTimestamp is the variable used to store either:
	// - creation timestamp, if the Cache's EvictionPolicy is FirstInFirstOut
	// - last access timestamp, if the Cache's EvictionPolicy is LeastRecentlyUsed
	RelevantTimestamp time.Time

	// Expiration is the unix time in nanoseconds at which the entry will expire (-1 means no expiration)
	Expiration int64

	next     *Entry
	previous *Entry
}

func (entry *Entry) Accessed() {
	entry.RelevantTimestamp = time.Now()
}

func (entry Entry) Expired() bool {
	if entry.Expiration > 0 {
		if time.Now().UnixNano() > entry.Expiration {
			return true
		}
	}
	return false
}
