package logger

import (
	"strings"
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// FileStream is the Stream that writes to a file
//   Any record with a level < FilterLevel will be written
type FileStream struct {
	*json.Encoder
	Path           string
	Converter      Converter
	FilterLevel    Level
	Unbuffered     bool
	file           *os.File
	output         *bufio.Writer
	flushFrequency time.Duration
	mutex          sync.Mutex
}

// SetFilterLevel sets the filter level
func (stream *FileStream) SetFilterLevel(level Level) Streamer {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	stream.FilterLevel = level
	return stream
}

// Write writes the given Record
//   implements logger.Stream
func (stream *FileStream) Write(record Record) (err error) {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	if stream.file == nil {
		const flags = os.O_CREATE | os.O_APPEND | os.O_WRONLY
		const perms = 0644
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
			stream.output  =  nil
			stream.Encoder = json.NewEncoder(stream.file)
		} else {
			stream.output  = bufio.NewWriter(stream.file)
			stream.Encoder = json.NewEncoder(stream.output)
			stream.flushFrequency = GetFlushFrequencyFromEnvironment()
			go stream.flushJob()
		}
	}
	if err := stream.Encoder.Encode(stream.Converter.Convert(record)); err != nil {
		return errors.WithStack(err)
	}
	if GetLevelFromRecord(record) >= ERROR && stream.output != nil {
		stream.output.Flush() // calling stream.Flush will Lock the mutex again and end up with a dead-lock
	}
	return nil
}

// ShouldWrite tells if the given level should be written to this stream
//   implements logger.Stream
func (stream *FileStream) ShouldWrite(level Level) bool {
	return level.ShouldWrite(stream.FilterLevel)
}

// Flush flushes the stream (makes sure records are actually written)
//   implements logger.Stream
func (stream *FileStream) Flush() {
	if stream.output != nil {
		stream.mutex.Lock()
		defer stream.mutex.Unlock()
		stream.output.Flush()
	}
}

// Close closes the stream
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
//   implements the fmt.Stringer interface
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