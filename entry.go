package gocache

import "time"

type Entry struct {
	Value interface{}

	// RelevantTimestamp is the variable used to store either:
	// - creation timestamp, if the Cache's EvictionPolicy is FirstInFirstOut
	// - last access timestamp, if the Cache's EvictionPolicy is LeastRecentlyUsed
	RelevantTimestamp time.Time
}

func (entry *Entry) Accessed() {
	entry.RelevantTimestamp = time.Now()
}
