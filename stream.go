package logger

import (
	"os"
	"strings"
	"time"
	"github.com/gildas/go-core"
)

// Streamer is the interface a Logger writes to
type Streamer interface {
	Write(record Record) error
	ShouldWrite(level Level) bool
	SetFilterLevel(level Level) Streamer
	Flush()
	Close()
}

// GetFlushFrequencyFromEnvironment fetches the flush frequency from the environment
//  the frequency should be like https://golang.org/pkg/time/#ParseDuration or an ISO8601 duration.  
//
//  If not set, the frequency will be 5 minutes
func GetFlushFrequencyFromEnvironment() time.Duration {
	return core.GetEnvAsDuration("LOG_FLUSHFREQUENCY", 5 * time.Minute)
}

// CreateStreamWithDestination creates a new Streamer from a list of strings
func CreateStreamWithDestination(destinations ...string) Streamer {
	unbuffered := false
	if value, ok := os.LookupEnv("DEBUG"); ok && value == "1" {
		unbuffered = true
	}
	if len(destinations) == 0 {
		destination, ok := os.LookupEnv("LOG_DESTINATION")
		if !ok || len(destination) == 0 {
			return &StdoutStream{Unbuffered: unbuffered}
		}
		destinations = []string{destination}
	}
	streams := []Streamer{}

	for _, destination := range destinations {
		var stream Streamer
		switch strings.ToLower(destination) {
		case "stdout":
			stream = &StdoutStream{Unbuffered: unbuffered}
		case "stderr":
			stream = &StderrStream{}
		case "gcp", "google", "googlecloud":
			stream = &StdoutStream{Unbuffered: true, Converter: &StackDriverConverter{}}
		case "stackdriver":
			stream =  &StackDriverStream{}
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
	return &MultiStream{ streams: streams }
}