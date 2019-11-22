package logger

import (
	"os"
	"strings"
	"time"
)

// Streamer is the interface a Logger writes to
type Streamer interface {
	Write(record Record) error
	ShouldWrite(level Level) bool
	Flush()
}

// GetFlushFrequencyFromEnvironment fetches the flush frequency from the environment
func GetFlushFrequencyFromEnvironment() time.Duration {
	if value, ok := os.LookupEnv("LOG_FLUSHFREQUENCY"); ok {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return 5 * time.Minute
}

// CreateStreamWithDestination creates a new Streamer from a list of strings
func CreateStreamWithDestination(destinations ...string) Streamer {
	if len(destinations) == 0 {
		destination, ok := os.LookupEnv("LOG_DESTINATION")
		if !ok || len(destination) == 0 {
			return &StdoutStream{}
		}
		destinations = []string{destination}
	}
	streams := []Streamer{}

	for _, destination := range destinations {
		var stream Streamer
		switch strings.ToLower(destination) {
		case "gcp", "google", "googlecloud":
			stream = &GCPStream{}
		case "stackdriver":
			stream =  &StackDriverStream{}
		case "nil", "null", "void", "blackhole", "nether":
			stream = &NilStream{}
		default:
			if strings.HasPrefix(destination, "file://") {
				stream = &FileStream{Path: strings.TrimPrefix(destination, "file://")}
			} else if len(destination) > 0 {
				stream = &FileStream{Path: destination}
			} else {
				stream = &StdoutStream{}
			}
		}
		streams = append(streams, stream)
	}
	if len(streams) == 1 {
		return streams[0]
	}
	return &MultiStream{ streams: streams }
}