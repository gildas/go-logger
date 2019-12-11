package logger

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateUnbufferedStreamWithDEBUG(t *testing.T) {
	os.Setenv("DEBUG", "1")
	defer os.Unsetenv("DEBUG")
	stream := CreateStreamWithDestination()
	assert.True(t, stream.(*StdoutStream).Unbuffered, "Stdout should be unbuffered with DEBUG set")
}

func TestFileStreamCanSetFilterLevel(t *testing.T) {
	stream := &FileStream{}
	assert.Equal(t, UNSET, stream.FilterLevel)
	stream.SetFilterLevel(WARN)
	assert.Equal(t, WARN, stream.FilterLevel)
}

func TestNilStreamCanSetFilterLevel(t *testing.T) {
	stream := &NilStream{}
	stream.SetFilterLevel(WARN)
}

func TestStderrStreamCanSetFilterLevel(t *testing.T) {
	stream := &StderrStream{}
	assert.Equal(t, UNSET, stream.FilterLevel)
	stream.SetFilterLevel(WARN)
	assert.Equal(t, WARN, stream.FilterLevel)
}

func TestStdoutStreamCanSetFilterLevel(t *testing.T) {
	stream := &StdoutStream{}
	assert.Equal(t, UNSET, stream.FilterLevel)
	stream.SetFilterLevel(WARN)
	assert.Equal(t, WARN, stream.FilterLevel)
}

func TestStackDriverStreamCanSetFilterLevel(t *testing.T) {
	stream := &StackDriverStream{}
	assert.Equal(t, UNSET, stream.FilterLevel)
	stream.SetFilterLevel(WARN)
	assert.Equal(t, WARN, stream.FilterLevel)
}

func TestMultiStreamCanSetFilterLevel(t *testing.T) {
	stream := &MultiStream{streams: []Streamer{&StdoutStream{}, &StderrStream{}}}
	assert.Equal(t, UNSET, stream.streams[0].(*StdoutStream).FilterLevel)
	assert.Equal(t, UNSET, stream.streams[1].(*StderrStream).FilterLevel)
	stream.SetFilterLevel(WARN)
	assert.Equal(t, WARN, stream.streams[0].(*StdoutStream).FilterLevel)
	assert.Equal(t, WARN, stream.streams[1].(*StderrStream).FilterLevel)
}