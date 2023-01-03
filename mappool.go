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
func (pool *MapPool) Get() (m map[string]interface{}) {
	return (*sync.Pool)(pool).Get().(map[string]interface{})
}

// Put adds a map to the Pool.
func (pool *MapPool) Put(m map[string]interface{}) {
	for key := range m {
		delete(m, key)
	}
	(*sync.Pool)(pool).Put(m)
}
