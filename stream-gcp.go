package logger

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/pkg/errors"
)


// GCPStream is the Stream that writes to the standard output
type GCPStream struct {
	*json.Encoder
	FilterLevel Level
	mutex       sync.Mutex
}

// SetFilterLevel sets the filter level
func (stream *GCPStream) SetFilterLevel(level Level) Streamer {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	stream.FilterLevel = level
	return stream
}

// Write writes the given Record
//   implements logger.Stream
func (stream *GCPStream) Write(record Record) error {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	if stream.Encoder == nil {
		stream.Encoder = json.NewEncoder(os.Stdout)
		if stream.FilterLevel == 0 {
			stream.FilterLevel = GetLevelFromEnvironment()
		}
	}
	record["severity"] = severity(record["level"]) // see stackdriver

	delete(record, "level")
	delete(record, "time")
	delete(record, "name")
	if err := stream.Encoder.Encode(record); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// ShouldWrite tells if the given level should be written to this stream
//   implements logger.Stream
func (stream *GCPStream) ShouldWrite(level Level) bool {
	return level.ShouldWrite(stream.FilterLevel)
}

// Flush flushes the stream (makes sure records are actually written)
//   implements logger.Stream
func (stream *GCPStream) Flush() {
}

// String gets a string version
//   implements the fmt.Stringer interface
func (stream *GCPStream) String() string {
	return "Stream to Google Cloud"
}