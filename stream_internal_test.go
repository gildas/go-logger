package logger

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateUnbufferedStreamWithDEBUG(t *testing.T) {
	currentDEBUG := os.Getenv("DEBUG")
	defer func() { _ = os.Setenv("DEBUG", currentDEBUG) }()

	_ = os.Setenv("DEBUG", "1")
	stream := CreateStream(NewLevelSet(DEBUG))
	assert.True(t, stream.(*StdoutStream).Unbuffered, "Stdout should be unbuffered with DEBUG set")
}

func TestMultiStreamCanSetFilterLevel(t *testing.T) {
	stream := &MultiStream{streams: []Streamer{&StdoutStream{}, &StderrStream{}}}
	assert.Equal(t, UNSET, stream.streams[0].(*StdoutStream).FilterLevels.GetDefault())
	assert.Equal(t, UNSET, stream.streams[1].(*StderrStream).FilterLevels.GetDefault())
	stream.SetFilterLevel(WARN)
	assert.Equal(t, WARN, stream.streams[0].(*StdoutStream).FilterLevels.GetDefault())
	assert.Equal(t, WARN, stream.streams[1].(*StderrStream).FilterLevels.GetDefault())
}
