package logger

import (
	"bufio"
	"encoding/json"
	"os"
	"time"

	"github.com/pkg/errors"
)


// StdoutStream is the Stream that writes to the standard output
type StdoutStream struct {
	*json.Encoder
	FilterLevel    Level
	Unbuffered     bool
	output         *bufio.Writer
	lastFlush      time.Time
	flushFrequency time.Duration
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
	if err := stream.Encoder.Encode(record); err != nil {
		return errors.WithStack(err)
	}
	// TODO: if the flush Frequency changed, restart the flushJob
	if GetLevelFromRecord(record) >= ERROR || time.Since(stream.lastFlush) >= stream.flushFrequency {
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
		stream.output.Flush()
	}
}

// String gets a string version
//   implements the fmt.Stringer interface
func (stream StdoutStream) String() string {
	if stream.Unbuffered {
		return "Unbuffered Stream to stdout"
	}
	return "Stream to stdout"
}

func (stream *StdoutStream) flushJob() {
	for range time.Tick(stream.flushFrequency) {
		// TODO: Need to be interruptible
		// TODO: Add lock
		stream.Flush()
		stream.lastFlush = time.Now()
	}
}