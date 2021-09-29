package logger

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gildas/go-errors"
)

// StdoutStream is the Stream that writes to the standard output
type StdoutStream struct {
	*json.Encoder
	Converter      Converter
	FilterLevel    Level
	Unbuffered     bool
	output         *bufio.Writer
	flushFrequency time.Duration
	mutex          sync.Mutex
}

// SetFilterLevel sets the filter level
func (stream *StdoutStream) SetFilterLevel(level Level) Streamer {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	stream.FilterLevel = level
	return stream
}

// SetFilterLevelIfUnset sets the filter level if not set already
func (stream *StdoutStream) SetFilterLevelIfUnset(level Level) Streamer {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	if stream.FilterLevel == UNSET {
		stream.FilterLevel = level
	}
	return stream
}

// Write writes the given Record
func (stream *StdoutStream) Write(record Record) error {
	// implements logger.Stream
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	if stream.Encoder == nil {
		if stream.FilterLevel == UNSET {
			stream.FilterLevel = GetLevelFromEnvironment()
		}
		if stream.Converter == nil {
			stream.Converter = GetConverterFromEnvironment()
		}
		if stream.Unbuffered {
			stream.output = nil
			stream.Encoder = json.NewEncoder(os.Stdout)
		} else {
			stream.output = bufio.NewWriter(os.Stdout)
			stream.Encoder = json.NewEncoder(stream.output)
			stream.flushFrequency = GetFlushFrequencyFromEnvironment()
			go stream.flushJob()
		}
	}
	if err := stream.Encoder.Encode(stream.Converter.Convert(record)); err != nil {
		return errors.JSONMarshalError.Wrap(err)
	}
	if GetLevelFromRecord(record) >= ERROR && stream.output != nil {
		stream.output.Flush() // calling stream.Flush would Lock the mutex again and end up with a dead-lock
	}
	return nil
}

// ShouldWrite tells if the given level should be written to this stream
func (stream *StdoutStream) ShouldWrite(level Level) bool {
	// implements logger.Stream
	return level.ShouldWrite(stream.FilterLevel)
}

// Flush flushes the stream (makes sure records are actually written)
func (stream *StdoutStream) Flush() {
	// implements logger.Stream
	if stream.output != nil {
		stream.mutex.Lock()
		defer stream.mutex.Unlock()
		stream.output.Flush()
	}
}

// Close closes the stream
func (stream *StdoutStream) Close() {
	if stream.output != nil {
		stream.mutex.Lock()
		defer stream.mutex.Unlock()
		stream.output.Flush()
	}
}

// String gets a string version
func (stream *StdoutStream) String() string {
	// implements the fmt.Stringer interface
	var format strings.Builder

	if stream.Unbuffered {
		format.WriteString("Unbuffered ")
	}
	format.WriteString("Stream to stdout")
	if stream.FilterLevel == UNSET {
		return format.String()
	}
	format.WriteString(", Filter: %s")
	return fmt.Sprintf(format.String(), stream.FilterLevel)
}

func (stream *StdoutStream) flushJob() {
	for range time.Tick(stream.flushFrequency) {
		stream.Flush()
	}
}
