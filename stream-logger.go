package logger

import (
	"fmt"
)

// Write writes the given Record
//
// implements logger.Streamer
func (log *Logger) Write(record *Record) error {
	// implements logger.Stream
	record.Merge(log.record)
	return log.stream.Write(record)
}

// GetFilterLevels gets the filter levels of the streamer
//
// implements logger.Streamer
func (log *Logger) GetFilterLevels() LevelSet {
	return log.stream.GetFilterLevels()
}

// SetFilterLevel sets the filter level of the streamer
//
// If present, the first parameter is the topic.
//
// If present, the second parameter is the scope.
//
// implements logger.FilterSetter
func (log *Logger) SetFilterLevel(level Level, parameters ...string) {
	if setter, ok := log.stream.(FilterSetter); ok {
		setter.SetFilterLevel(level, parameters...)
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
func (log *Logger) FilterMore() {
	if modifier, ok := log.stream.(FilterModifier); ok {
		modifier.FilterMore()
	}
}

// FilterLess tells the stream to filter less
//
// The stream will filter less if it is not already at the lowest level.
// Which means more log messages will be written to the stream
//
// Example: if the stream is at INFO, it will be filtering at DEBUG
//
// implements logger.FilterModifier
func (log *Logger) FilterLess() {
	if modifier, ok := log.stream.(FilterModifier); ok {
		modifier.FilterLess()
	}
}

// ShouldLogSourceInfo tells if the source info should be logged
//
// implements logger.Streamer
func (log *Logger) ShouldLogSourceInfo() bool {
	return log.stream.ShouldLogSourceInfo()
}

// ShouldWrite tells if the given level should be written to this stream
//
// implements logger.Streamer
func (log *Logger) ShouldWrite(level Level, topic, scope string) bool {
	return log.stream.ShouldWrite(level, topic, scope)
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
//	implements fmt.Stringer
func (log Logger) String() string {
	return fmt.Sprintf("Logger(%s)", log.stream)
}
