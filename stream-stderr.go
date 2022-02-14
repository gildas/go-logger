package logger

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/gildas/go-errors"
)

// StderrStream is the Stream that writes to the standard error
type StderrStream struct {
	*json.Encoder
	Converter   Converter
	FilterLevel Level
	mutex       sync.Mutex
}

// SetFilterLevel sets the filter level
//
// implements logger.FilterSetter
func (stream *StderrStream) SetFilterLevel(level Level) {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	stream.FilterLevel = level
}

// SetFilterLevelIfUnset sets the filter level if not set already
//
// implements logger.FilterSetter
func (stream *StderrStream) SetFilterLevelIfUnset(level Level) {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	if stream.FilterLevel == UNSET {
		stream.FilterLevel = level
	}
}

// Write writes the given Record
//
// implements logger.Streamer
func (stream *StderrStream) Write(record Record) error {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	if stream.Encoder == nil {
		stream.Encoder = json.NewEncoder(os.Stderr)
		if stream.FilterLevel == UNSET {
			stream.FilterLevel = GetLevelFromEnvironment()
		}
	}
	if stream.Converter == nil {
		stream.Converter = GetConverterFromEnvironment()
	}
	if err := stream.Encoder.Encode(stream.Converter.Convert(record)); err != nil {
		return errors.JSONMarshalError.Wrap(err)
	}
	return nil
}

// ShouldWrite tells if the given level should be written to this stream
//
// implements logger.Streamer
func (stream *StderrStream) ShouldWrite(level Level) bool {
	return level.ShouldWrite(stream.FilterLevel)
}

// Flush flushes the stream (makes sure records are actually written)
//
// implements logger.Streamer
func (stream *StderrStream) Flush() {
}

// Close closes the stream
//
// implements logger.Streamer
func (stream *StderrStream) Close() {
}

// String gets a string version
//
// implements fmt.Stringer
func (stream *StderrStream) String() string {
	return "Stream to stderr"
}
