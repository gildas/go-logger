package logger

import "github.com/gildas/go-errors"

// MultiStream is the Stream that writes to several streams
type MultiStream struct {
	streams []Streamer
}

// GetFilterLevels gets the filter levels
//
// # If the multi stream contains at least one stream, it returns the filter levels of the first stream
//
// implements logger.Streamer
func (stream *MultiStream) GetFilterLevels() LevelSet {
	if len(stream.streams) == 0 {
		return LevelSet{}
	}
	return stream.streams[0].GetFilterLevels()
}

// SetFilterLevel sets the filter level
//
// If present, the first parameter is the topic.
//
// If present, the second parameter is the scope.
//
// implements logger.FilterSetter
func (stream *MultiStream) SetFilterLevel(level Level, parameters ...string) {
	for _, s := range stream.streams {
		if setter, ok := s.(FilterSetter); ok {
			setter.SetFilterLevel(level, parameters...)
		}
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
func (stream *MultiStream) FilterMore() {
	for _, s := range stream.streams {
		if modifier, ok := s.(FilterModifier); ok {
			modifier.FilterMore()
		}
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
func (stream *MultiStream) FilterLess() {
	for _, s := range stream.streams {
		if modifier, ok := s.(FilterModifier); ok {
			modifier.FilterLess()
		}
	}
}

// ShouldLogSourceInfo tells if the source info should be logged
//
// # If at least one stream returns true, the stream should write the record
//
// implements logger.Streamer
func (stream *MultiStream) ShouldLogSourceInfo() bool {
	for _, s := range stream.streams {
		if s.ShouldLogSourceInfo() {
			return true
		}
	}
	return false
}

// Write writes the given Record
//
// implements logger.Streamer
func (stream *MultiStream) Write(record *Record) error {
	var errs errors.MultiError

	for _, s := range stream.streams {
		if err := s.Write(record); err != nil {
			errs.Append(err)
		}
	}
	return errs.AsError()
}

// ShouldWrite tells if the given level should be written to this stream
//
// # If at least one stream returns true, the stream should write the record
//
// implements logger.Streamer
func (stream *MultiStream) ShouldWrite(level Level, topic, scope string) bool {
	for _, s := range stream.streams {
		if s.ShouldWrite(level, topic, scope) {
			return true
		}
	}
	return false
}

// Flush flushes the stream (makes sure records are actually written)
//
// implements logger.Streamer
func (stream *MultiStream) Flush() {
	for _, s := range stream.streams {
		s.Flush()
	}
}

// Close closes the stream
//
// implements logger.Streamer
func (stream *MultiStream) Close() {
	for _, s := range stream.streams {
		s.Close()
	}
}

// String gets a string version
//
// implements fmt.Stringer
func (stream MultiStream) String() string {
	return "MultiStream"
}

// CreateMultiStream creates a MultiStream that contains all given Streamer objects
func CreateMultiStream(streams ...Streamer) Streamer {
	return &MultiStream{streams: streams}
}
