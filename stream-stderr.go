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
	Converter    Converter
	FilterLevel  Level
	FilterLevels TopicScopeLevels
	mutex        sync.Mutex
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

// SetFilterLevelForTopic sets the filter level for a given topic
//
// implements logger.FilterSetter
func (stream *StderrStream) SetFilterLevelForTopic(level Level, topic string) {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	stream.FilterLevels.Set(topic, "", level)
}

// SetFilterLevelForTopicAndScope sets the filter level for a given topic
//
// implements logger.FilterSetter
func (stream *StderrStream) SetFilterLevelForTopicAndScope(level Level, topic, scope string) {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	stream.FilterLevels.Set(topic, scope, level)
}

// FilterMore tells the stream to filter more
//
// The stream will filter more if it is not already at the highest level.
// Which means less log messages will be written to the stream
//
// Example: if the stream is at DEBUG, it will be filtering at INFO
//
// implements logger.FilterModifier
func (stream *StderrStream) FilterMore() {
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
func (stream *StderrStream) FilterLess() {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	stream.FilterLevel = stream.FilterLevel.Previous()
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
func (stream *StderrStream) ShouldWrite(level Level, topic, scope string) bool {
	if _level, found := stream.FilterLevels.Get(topic, scope); found {
		return level.ShouldWrite(_level)
	}
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
