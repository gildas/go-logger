package logger

import "sync"

type MapPool sync.Pool

var mapPool = NewMapPool()

// NewMapPool creates a new MapPool
func NewMapPool() *MapPool {
	return &MapPool{
		New: func() interface{} {
			return make(map[string]interface{})
		},
	}
}

// Get selects an arbitrary map from the Pool, removes it from the Pool, and returns it to the caller.
func (pool *MapPool) Get() (*Record) {
	return (*sync.Pool)(pool).Get().(*Record)
}

// Put adds a map to the Pool.
func (pool *MapPool) Put(record *Record) {
	record.Reset()
	(*sync.Pool)(pool).Put(record)
}
