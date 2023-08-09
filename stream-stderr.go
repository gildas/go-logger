package logger

import (
	"fmt"
	"os"
	"sync"

	"github.com/gildas/go-errors"
)

// StderrStream is the Stream that writes to the standard error
type StderrStream struct {
	Converter    Converter
	FilterLevels LevelSet
	SourceInfo   bool
	mutex        sync.Mutex
}

// SetFilterLevel sets the filter level
//
// If present, the first parameter is the topic.
//
// If present, the second parameter is the scope.
//
// implements logger.FilterSetter
func (stream *StderrStream) SetFilterLevel(level Level, parameters ...string) {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	if len(parameters) == 0 {
		stream.FilterLevels.SetDefault(level)
	} else if len(parameters) == 1 {
		stream.FilterLevels.Set(level, parameters[0], "")
	} else {
		stream.FilterLevels.Set(level, parameters[0], parameters[1])
	}
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
	stream.FilterLevels.SetDefault(stream.FilterLevels.GetDefault().Next())
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
	stream.FilterLevels.SetDefault(stream.FilterLevels.GetDefault().Previous())
}

// Write writes the given Record
//
// implements logger.Streamer
func (stream *StderrStream) Write(record Record) (err error) {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	if stream.Converter == nil {
		stream.Converter = GetConverterFromEnvironment()
	}
	if len(stream.FilterLevels) == 0 {
		stream.FilterLevels = ParseLevelsFromEnvironment()
	}
	payload, _ := stream.Converter.Convert(record).MarshalJSON()
	if _, err = os.Stderr.Write(payload); err != nil {
		return errors.WithStack(err)
	}
	if _, err = os.Stderr.Write([]byte("\n")); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// ShouldLogSourceInfo tells if the source info should be logged
//
// implements logger.Streamer
func (stream *StderrStream) ShouldLogSourceInfo() bool {
	return stream.SourceInfo
}

// ShouldWrite tells if the given level should be written to this stream
//
// implements logger.Streamer
func (stream *StderrStream) ShouldWrite(level Level, topic, scope string) bool {
	return level.ShouldWrite(stream.FilterLevels.Get(topic, scope))
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
	if len(stream.FilterLevels) > 0 {
		return fmt.Sprintf("Stream to stderr, Filter: %s", stream.FilterLevels)
	}
	return "Stream to stderr"
}
