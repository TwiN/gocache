package gocache

import "time"

type Entry struct {
	Key   string
	Value interface{}

	// RelevantTimestamp is the variable used to store either:
	// - creation timestamp, if the Cache's EvictionPolicy is FirstInFirstOut
	// - last access timestamp, if the Cache's EvictionPolicy is LeastRecentlyUsed
	RelevantTimestamp time.Time

	next     *Entry
	previous *Entry
}

func (entry *Entry) Accessed() {
	entry.RelevantTimestamp = time.Now()
}
