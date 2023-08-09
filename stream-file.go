package logger

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"sync"
	"time"

	"github.com/gildas/go-errors"
)

// FileStream is the Stream that writes to a file
//
//	Any record with a level < FilterLevel will be written
type FileStream struct {
	Path           string
	Converter      Converter
	FilterLevels   LevelSet
	Unbuffered     bool
	SourceInfo     bool
	file           *os.File
	output         *bufio.Writer
	writer         io.Writer
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
func (stream *FileStream) SetFilterLevel(level Level, parameters ...string) {
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
func (stream *FileStream) FilterMore() {
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
func (stream *FileStream) FilterLess() {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	stream.FilterLevels.SetDefault(stream.FilterLevels.GetDefault().Previous())
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
		if stream.Converter == nil {
			stream.Converter = GetConverterFromEnvironment()
		}
		if len(stream.FilterLevels) == 0 {
			stream.FilterLevels = ParseLevelsFromEnvironment()
		}
		if stream.Unbuffered {
			stream.output = nil
			stream.writer = stream.file
		} else {
			stream.output = bufio.NewWriter(stream.file)
			stream.writer = stream.output
			stream.flushFrequency = GetFlushFrequencyFromEnvironment()
			go stream.flushJob()
		}
	}
	payload, _ := stream.Converter.Convert(record).MarshalJSON()
	if _, err = stream.writer.Write(payload); err != nil {
		return errors.WithStack(err)
	}
	if _, err = stream.writer.Write([]byte("\n")); err != nil {
		return errors.WithStack(err)
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
	return level.ShouldWrite(stream.FilterLevels.Get(topic, scope))
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
	format := bufferPool.Get()
	defer bufferPool.Put(format)

	if stream.Unbuffered {
		format.WriteString("Unbuffered ")
	}
	format.WriteString("Stream to %s")
	if len(stream.FilterLevels) > 0 {
		format.WriteString(", Filter: %s")
		return fmt.Sprintf(format.String(), stream.Path, stream.FilterLevels)
	}
	return fmt.Sprintf(format.String(), stream.Path)
}

func (stream *FileStream) flushJob() {
	for range time.Tick(stream.flushFrequency) {
		stream.Flush()
	}
}
