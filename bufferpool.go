package logger

import (
	"bytes"
	"sync"
)

type BufferPool sync.Pool

var bufferPool = NewBufferPool()

// NewBufferPool creates a new BufferPool
func NewBufferPool() *BufferPool {
	return &BufferPool{
		New: func() any {
			return &bytes.Buffer{}
		},
	}
}

// Get selects an arbitrary Buffer from the Pool, removes it from the Pool, and returns it to the caller.
func (pool *BufferPool) Get() (buffer *bytes.Buffer) {
	return (*sync.Pool)(pool).Get().(*bytes.Buffer)
}

// Put adds a Buffer to the Pool.
func (pool *BufferPool) Put(buffer *bytes.Buffer) {
	buffer.Reset()
	(*sync.Pool)(pool).Put(buffer)
}
