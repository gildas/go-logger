package logger

import (
	"os"
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