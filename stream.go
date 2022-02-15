package logger

import (
	"os"
	"strings"
	"time"

	"github.com/gildas/go-core"
)

// Streamer is the interface a Logger writes to
type Streamer interface {
	// Write writes the given Record
	Write(record Record) error

	// ShouldWrite tells if the given level should be written to this stream
	ShouldWrite(level Level) bool

	// ShouldWriteWithTopic tells if the given level should be written to this stream
	ShouldWriteWithTopic(level Level, topic string) bool

	// ShouldWriteWithTopicAndScope tells if the given level should be written to this stream
	ShouldWriteWithTopicAndScope(level Level, topic, scope string) bool

	// Flush flushes the stream (makes sure records are actually written)
	Flush()

	// Close closes the stream
	Close()
}

// GetFlushFrequencyFromEnvironment fetches the flush frequency from the environment
//
// the frequency should be like https://golang.org/pkg/time/#ParseDuration or an ISO8601 duration.
//
// If not set, the frequency will be 5 minutes
func GetFlushFrequencyFromEnvironment() time.Duration {
	return core.GetEnvAsDuration("LOG_FLUSHFREQUENCY", 5*time.Minute)
}

// CreateStreamWithDestination creates a new Streamer from a list of strings
//
// "stdout" will create a StdoutStream
//
// "stderr" will create a StderrStream
//
// "nil", "null" will create a NilStream
//
// "stackdriver" will create a StackDriverStream
//
// "gcp", "googlecloud", "google" will create a StdoutStream, unbuffered, with the StackDriverConverter
//
// "file:///path/to/file" or "path/to/file", "/path/to/file" will create a FileStream on the given location
//
// If more than one string is given, a MultiStream of all Streams from strings is created.
//
// If the environment variable DEBUG is set to 1, all Streams are created unbuffered.
//
// If the list is empty, the environment variable LOG_DESTINATION is used.
func CreateStreamWithDestination(destinations ...string) Streamer {
	unbuffered := core.GetEnvAsBool("DEBUG", false)
	if len(destinations) == 0 {
		destination, ok := os.LookupEnv("LOG_DESTINATION")
		if !ok || len(destination) == 0 {
			return &StdoutStream{Unbuffered: unbuffered}
		}
		destinations = strings.Split(destination, ",")
	}
	streams := []Streamer{}

	for _, destination := range destinations {
		var stream Streamer
		switch strings.ToLower(strings.TrimSpace(destination)) {
		case "stdout":
			stream = &StdoutStream{Unbuffered: unbuffered}
		case "stderr":
			stream = &StderrStream{}
		case "gcp", "google", "googlecloud":
			stream = &StdoutStream{Unbuffered: true, Converter: &StackDriverConverter{}}
		case "stackdriver":
			stream = &StackDriverStream{}
		case "nil", "null", "void", "blackhole", "nether":
			stream = &NilStream{}
		default:
			if strings.HasPrefix(destination, "file://") {
				stream = &FileStream{Path: strings.TrimPrefix(destination, "file://"), Unbuffered: unbuffered}
			} else if len(destination) > 0 {
				stream = &FileStream{Path: destination, Unbuffered: unbuffered}
			} else {
				stream = &StdoutStream{Unbuffered: unbuffered}
			}
		}
		streams = append(streams, stream)
	}
	if len(streams) == 1 {
		return streams[0]
	}
	return &MultiStream{streams: streams}
}
