package logger

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/pkg/errors"
)


// StderrStream is the Stream that writes to the standard output
type StderrStream struct {
	*json.Encoder
	Converter   Converter
	FilterLevel Level
	mutex       sync.Mutex
}

// SetFilterLevel sets the filter level
func (stream *StderrStream) SetFilterLevel(level Level) Streamer {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	stream.FilterLevel = level
	return stream
}

// Write writes the given Record
//   implements logger.Stream
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
		return errors.WithStack(err)
	}
	return nil
}

// ShouldWrite tells if the given level should be written to this stream
//   implements logger.Stream
func (stream *StderrStream) ShouldWrite(level Level) bool {
	return level.ShouldWrite(stream.FilterLevel)
}

// Flush flushes the stream (makes sure records are actually written)
//   implements logger.Stream
func (stream *StderrStream) Flush() {
}

// Close closes the stream
func (stream *StderrStream) Close() {
}

// String gets a string version
//   implements the fmt.Stringer interface
func (stream *StderrStream) String() string {
	return "Stream to stderr"
}