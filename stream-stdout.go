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
	FilterLevels   map[string]Level
	Unbuffered     bool
	output         *bufio.Writer
	flushFrequency time.Duration
	mutex          sync.Mutex
}

// SetFilterLevel sets the filter level
//
// implements logger.FilterSetter
func (stream *StdoutStream) SetFilterLevel(level Level) {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	stream.FilterLevel = level
}

// SetFilterLevelIfUnset sets the filter level if not set already
//
// implements logger.FilterSetter
func (stream *StdoutStream) SetFilterLevelIfUnset(level Level) {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	if stream.FilterLevel == UNSET {
		stream.FilterLevel = level
	}
}

// SetFilterLevelForTopic sets the filter level for a given topic
//
// implements logger.FilterSetter
func (stream *StdoutStream) SetFilterLevelForTopic(level Level, topic string) {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	if stream.FilterLevels == nil {
		stream.FilterLevels = make(map[string]Level)
	}
	stream.FilterLevels[topic] = level
}

// SetFilterLevelForTopicAndScope sets the filter level for a given topic
//
// implements logger.FilterSetter
func (stream *StdoutStream) SetFilterLevelForTopicAndScope(level Level, topic, scope string) {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	if stream.FilterLevels == nil {
		stream.FilterLevels = make(map[string]Level)
	}
	stream.FilterLevels[topic + "|" + scope] = level
}

// FilterMore tells the stream to filter more
//
// The stream will filter more if it is not already at the highest level.
// Which means less log messages will be written to the stream
//
// Example: if the stream is at DEBUG, it will be filtering at INFO
//
// implements logger.FilterModifier
func (stream *StdoutStream) FilterMore() {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	stream.FilterLevel = stream.FilterLevel.Next()
}

// FilterLess tells the stream to filter less
//
// The stream will filter less if it is not already at the lowest level.
// Which means more log messages will be written to the stream
//
// Example: if the stream is at INFO, it will be filtering at DEBUG
//
// implements logger.FilterModifier
func (stream *StdoutStream) FilterLess() {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	stream.FilterLevel = stream.FilterLevel.Previous()
}

// Write writes the given Record
//
// implements logger.Streamer
func (stream *StdoutStream) Write(record Record) error {
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
//
// implements logger.Streamer
func (stream *StdoutStream) ShouldWrite(level Level) bool {
	return level.ShouldWrite(stream.FilterLevel)
}

// ShouldWriteWithTopic tells if the given level should be written to this stream
//
// implements logger.Streamer
func (stream *StdoutStream) ShouldWriteWithTopic(level Level, topic string) bool {
	if _level, found := stream.FilterLevels[topic]; found {
		return level.ShouldWrite(_level)
	}
	return level.ShouldWrite(stream.FilterLevel)
}

// ShouldWriteWithTopicAndScope tells if the given level should be written to this stream
//
// implements logger.Streamer
func (stream *StdoutStream) ShouldWriteWithTopicAndScope(level Level, topic, scope string) bool {
	if _level, found := stream.FilterLevels[topic + "|" + scope]; found {
		return level.ShouldWrite(_level)
	}
	if _level, found := stream.FilterLevels[topic]; found {
		return level.ShouldWrite(_level)
	}
	return level.ShouldWrite(stream.FilterLevel)
}

// Flush flushes the stream (makes sure records are actually written)
//
// implements logger.Streamer
func (stream *StdoutStream) Flush() {
	if stream.output != nil {
		stream.mutex.Lock()
		defer stream.mutex.Unlock()
		stream.output.Flush()
	}
}

// Close closes the stream
//
// implements logger.Streamer
func (stream *StdoutStream) Close() {
	if stream.output != nil {
		stream.mutex.Lock()
		defer stream.mutex.Unlock()
		stream.output.Flush()
	}
}

// String gets a string version
//
// implements fmt.Stringer
func (stream *StdoutStream) String() string {
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
