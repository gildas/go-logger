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
	Write(record *Record) error

	// ShouldWrite tells if the given level should be written to this stream
	ShouldWrite(level Level, topic, scope string) bool

	// ShouldLogSourceInfo tells if the source info should be logged
	ShouldLogSourceInfo() bool

	// Flush flushes the stream (makes sure records are actually written)
	Flush()

	// Close closes the stream
	Close()

	// GetFilterLevels gets the FilterLevels
	GetFilterLevels() LevelSet
}

// GetFlushFrequencyFromEnvironment fetches the flush frequency from the LOG_FLUSHFREQUENCY environment
//
// the frequency should be like https://golang.org/pkg/time/#ParseDuration or an ISO8601 duration.
//
// If not set, the frequency will be 5 minutes
func GetFlushFrequencyFromEnvironment() time.Duration {
	return GetFlushFrequencyFromEnvironmentWithPrefix("")
}

// GetFlushFrequencyFromEnvironmentWithPrefix fetches the flush frequency from the environment
// using the given prefix for the LOG_FLUSHFREQUENCY environment variable.
//
// the frequency should be like https://golang.org/pkg/time/#ParseDuration or an ISO8601 duration.
//
// If not set, the frequency will be 5 minutes
func GetFlushFrequencyFromEnvironmentWithPrefix(prefix EnvironmentPrefix) time.Duration {
	return core.GetEnvAsDuration(string(prefix)+"LOG_FLUSHFREQUENCY", 5*time.Minute)
}

// CreateStream creates a new Streamer from a list of strings
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
func CreateStream(levels LevelSet, destinations ...string) Streamer {
	return CreateStreamWithPrefix("", levels, destinations...)
}

// CreateStreamWithPrefix creates a new Streamer from a list of strings
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
func CreateStreamWithPrefix(prefix EnvironmentPrefix, levels LevelSet, destinations ...string) Streamer {
	debug := levels.Get("any", "any") == DEBUG
	unbuffered := debug
	sourceInfo := core.GetEnvAsBool(string(prefix)+"LOG_SOURCEINFO", false)
	if len(destinations) == 0 {
		destination, ok := os.LookupEnv(string(prefix) + "LOG_DESTINATION")
		if !ok || len(destination) == 0 {
			return &StdoutStream{FilterLevels: levels, Unbuffered: unbuffered, SourceInfo: sourceInfo}
		}
		destinations = strings.Split(destination, ",")
	}
	streams := []Streamer{}

	for _, destination := range destinations {
		var stream Streamer
		switch strings.ToLower(strings.TrimSpace(destination)) {
		case "stdout":
			stream = &StdoutStream{FilterLevels: levels, Unbuffered: unbuffered, SourceInfo: sourceInfo, environmentPrefix: prefix}
		case "stderr":
			stream = &StderrStream{FilterLevels: levels, SourceInfo: sourceInfo, environmentPrefix: prefix}
		case "gcp", "google", "googlecloud":
			stream = &StdoutStream{FilterLevels: levels, Unbuffered: true, SourceInfo: sourceInfo, Converter: &StackDriverConverter{}, environmentPrefix: prefix}
		case "stackdriver":
			stream = &StackDriverStream{FilterLevels: levels, SourceInfo: sourceInfo}
		case "nil", "null", "void", "blackhole", "nether":
			stream = &NilStream{}
		default:
			if strings.HasPrefix(destination, "file://") {
				stream = &FileStream{FilterLevels: levels, Path: strings.TrimPrefix(destination, "file://"), Unbuffered: unbuffered, environmentPrefix: prefix, SourceInfo: sourceInfo}
			} else if len(destination) > 0 {
				stream = &FileStream{FilterLevels: levels, Path: destination, Unbuffered: unbuffered, SourceInfo: sourceInfo, environmentPrefix: prefix}
			} else {
				stream = &StdoutStream{FilterLevels: levels, Unbuffered: unbuffered, SourceInfo: sourceInfo, environmentPrefix: prefix}
			}
		}
		streams = append(streams, stream)
	}
	if len(streams) == 1 {
		return streams[0]
	}
	return &MultiStream{streams: streams}
}
