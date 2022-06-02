package logger

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/gildas/go-errors"
)

// FileStream is the Stream that writes to a file
//   Any record with a level < FilterLevel will be written
type FileStream struct {
	*json.Encoder
	Path           string
	Converter      Converter
	FilterLevel    Level
	FilterLevels   TopicScopeLevels
	Unbuffered     bool
	SourceInfo     bool
	file           *os.File
	output         *bufio.Writer
	flushFrequency time.Duration
	mutex          sync.Mutex
}

// SetFilterLevel sets the filter level
//
// implements logger.FilterSetter
func (stream *FileStream) SetFilterLevel(level Level) {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	stream.FilterLevel = level
}

// SetFilterLevelIfUnset sets the filter level if not set already
//
// implements logger.FilterSetter
func (stream *FileStream) SetFilterLevelIfUnset(level Level) {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	if stream.FilterLevel == UNSET {
		stream.FilterLevel = level
	}
}

// SetFilterLevelForTopic sets the filter level for a given topic
//
// implements logger.FilterSetter
func (stream *FileStream) SetFilterLevelForTopic(level Level, topic string) {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	stream.FilterLevels.Set(topic, "", level)
}

// SetFilterLevelForTopicAndScope sets the filter level for a given topic
//
// implements logger.FilterSetter
func (stream *FileStream) SetFilterLevelForTopicAndScope(level Level, topic, scope string) {
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
func (stream *FileStream) FilterMore() {
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
func (stream *FileStream) FilterLess() {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	stream.FilterLevel = stream.FilterLevel.Previous()
}

// Write writes the given Record
//
// implements logger.Streamer
func (stream *FileStream) Write(record Record) (err error) {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	if stream.file == nil {
		const flags = os.O_CREATE | os.O_APPEND | os.O_WRONLY
		const perms = 0644
		err = os.MkdirAll(path.Dir(stream.Path), os.ModePerm)
		if err != nil {
			return errors.WithStack(err)
		}
		if stream.file, err = os.OpenFile(stream.Path, flags, perms); err != nil {
			return errors.WithStack(err)
		}
		if stream.FilterLevel == UNSET {
			stream.FilterLevel = GetLevelFromEnvironment()
		}
		if stream.Converter == nil {
			stream.Converter = GetConverterFromEnvironment()
		}
		if stream.Unbuffered {
			stream.output = nil
			stream.Encoder = json.NewEncoder(stream.file)
		} else {
			stream.output = bufio.NewWriter(stream.file)
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
func (stream *FileStream) ShouldWrite(level Level, topic, scope string) bool {
	if _level, found := stream.FilterLevels.Get(topic, scope); found {
		return level.ShouldWrite(_level)
	}
	return level.ShouldWrite(stream.FilterLevel)
}

// ShouldLogSourceInfo tells if the source info should be logged
//
// implements logger.Streamer
func (stream *FileStream) ShouldLogSourceInfo() bool {
	return stream.SourceInfo
}

// Flush flushes the stream (makes sure records are actually written)
//
// implements logger.Streamer
func (stream *FileStream) Flush() {
	if stream.output != nil {
		stream.mutex.Lock()
		defer stream.mutex.Unlock()
		stream.output.Flush()
	}
}

// Close closes the stream
//
// implements logger.Streamer
func (stream *FileStream) Close() {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	if stream.output != nil {
		stream.output.Flush()
	}
	if stream.file != nil {
		stream.file.Close()
	}
}

// String gets a string version
//
// implements fmt.Stringer
func (stream *FileStream) String() string {
	var format strings.Builder

	if stream.Unbuffered {
		format.WriteString("Unbuffered ")
	}
	format.WriteString("Stream to %s")
	if stream.FilterLevel == UNSET {
		return fmt.Sprintf(format.String(), stream.Path)
	}
	format.WriteString(", Filter: %s")
	return fmt.Sprintf(format.String(), stream.Path, stream.FilterLevel)
}

func (stream *FileStream) flushJob() {
	for range time.Tick(stream.flushFrequency) {
		stream.Flush()
	}
}
