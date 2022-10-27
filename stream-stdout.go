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
	FilterLevels   LevelSet
	Unbuffered     bool
	SourceInfo     bool
	output         *bufio.Writer
	flushFrequency time.Duration
	mutex          sync.Mutex
}

// SetFilterLevel sets the filter level
//
// If present, the first parameter is the topic.
//
// If present, the second parameter is the scope.
//
// implements logger.FilterSetter
func (stream *StdoutStream) SetFilterLevel(level Level, parameters ...string) {
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
func (stream *StdoutStream) FilterMore() {
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
func (stream *StdoutStream) FilterLess() {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	stream.FilterLevels.SetDefault(stream.FilterLevels.GetDefault().Previous())
}

// Write writes the given Record
//
// implements logger.Streamer
func (stream *StdoutStream) Write(record Record) error {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	if stream.Encoder == nil {
		if stream.Converter == nil {
			stream.Converter = GetConverterFromEnvironment()
		}
		if len(stream.FilterLevels) == 0 {
			stream.FilterLevels = ParseLevelsFromEnvironment()
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
	if err := stream.Encoder.Encode(stream.Converter.Convert(record)); errors.Is(err, errors.JSONMarshalError) {
		return err
	} else if err != nil {
		return errors.JSONMarshalError.Wrap(err)
	}
	if GetLevelFromRecord(record) >= ERROR && stream.output != nil {
		stream.output.Flush() // calling stream.Flush would Lock the mutex again and end up with a dead-lock
	}
	return nil
}

// ShouldLogSourceInfo tells if the source info should be logged
//
// implements logger.Streamer
func (stream *StdoutStream) ShouldLogSourceInfo() bool {
	return stream.SourceInfo
}

// ShouldWrite tells if the given level should be written to this stream
//
// implements logger.Streamer
func (stream *StdoutStream) ShouldWrite(level Level, topic, scope string) bool {
	return level.ShouldWrite(stream.FilterLevels.Get(topic, scope))
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
	if len(stream.FilterLevels) > 0 {
		format.WriteString(", Filter: %s")
		return fmt.Sprintf(format.String(), stream.FilterLevels)
	}
	return format.String()
}

func (stream *StdoutStream) flushJob() {
	for range time.Tick(stream.flushFrequency) {
		stream.Flush()
	}
}
