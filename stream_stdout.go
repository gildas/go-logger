package logger

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
)


// StdoutStream is the Stream that writes to the standard output
type StdoutStream struct {
	*json.Encoder
	FilterLevel    Level
	Unbuffered     bool
	output         *bufio.Writer
	flushFrequency time.Duration
	mutex          sync.Mutex
}

// Write writes the given Record
//   implements logger.Stream
func (stream *StdoutStream) Write(record Record) error {
	if stream.Encoder == nil {
		if stream.FilterLevel == UNSET {
			stream.FilterLevel = GetLevelFromEnvironment()
		}
		if stream.Unbuffered {
			stream.output =  nil
			stream.Encoder = json.NewEncoder(os.Stdout)
		} else {
			stream.output = bufio.NewWriter(os.Stdout)
			stream.Encoder = json.NewEncoder(stream.output)
			stream.flushFrequency = GetFlushFrequencyFromEnvironment()
			go stream.flushJob()
		}
	}
	{
		stream.mutex.Lock()
		defer stream.mutex.Unlock()
		if err := stream.Encoder.Encode(record); err != nil {
			return errors.WithStack(err)
		}
	}
	if GetLevelFromRecord(record) >= ERROR {
		stream.Flush()
	}
	return nil
}

// ShouldWrite tells if the given level should be written to this stream
//   implements logger.Stream
func (stream *StdoutStream) ShouldWrite(level Level) bool {
	return level.ShouldWrite(stream.FilterLevel)
}

// Flush flushes the stream (makes sure records are actually written)
//   implements logger.Stream
func (stream *StdoutStream) Flush() {
	if stream.output != nil {
		stream.mutex.Lock()
		defer stream.mutex.Unlock()
		stream.output.Flush()
	}
}

// String gets a string version
//   implements the fmt.Stringer interface
func (stream *StdoutStream) String() string {
	if stream.Unbuffered {
		return "Unbuffered Stream to stdout"
	}
	return "Stream to stdout"
}

func (stream *StdoutStream) flushJob() {
	for range time.Tick(stream.flushFrequency) {
		stream.Flush()
	}
}