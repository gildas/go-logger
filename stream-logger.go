package logger

import (
	"fmt"
)

// Write writes the given Record
//
// implements logger.Streamer
func (log *Logger) Write(record Record) error {
	// implements logger.Stream
	record.Merge(log.record)
	return log.stream.Write(record)
}

// SetFilterLevel sets the filter level of the streamer
//
// implements logger.Streamer
func (log *Logger) SetFilterLevel(level Level) Streamer {
	log.stream.SetFilterLevel(level)
	return log.stream
}

// SetFilterLevelIfUnset sets the filter level if not set already
//
// implements logger.Streamer
func (log *Logger) SetFilterLevelIfUnset(level Level) Streamer {
	log.stream.SetFilterLevelIfUnset(level)
	return log.stream
}

// ShouldWrite tells if the given level should be written to this stream
//
// implements logger.Streamer
func (log *Logger) ShouldWrite(level Level) bool {
	return log.stream.ShouldWrite(level)
}

// Flush flushes the stream (makes sure records are actually written)
//
// implements logger.Streamer
func (log *Logger) Flush() {
	log.stream.Flush()
}

// Close closes the logger's stream
//
// implements logger.Streamer
func (log *Logger) Close() {
	log.stream.Close()
}

// String gets a string version
//
//   implements fmt.Stringer
func (log Logger) String() string {
	return fmt.Sprintf("Logger(%s)", log.stream)
}
