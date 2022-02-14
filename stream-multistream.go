package logger

import "github.com/gildas/go-errors"

// MultiStream is the Stream that writes to several streams
type MultiStream struct {
	streams []Streamer
}

// SetFilterLevel sets the filter level of all streams
//
// implements logger.FilterSetter
func (stream *MultiStream) SetFilterLevel(level Level) {
	for _, s := range stream.streams {
		if setter, ok := s.(FilterSetter); ok {
			setter.SetFilterLevel(level)
		}
	}
}

// SetFilterLevelIfUnset sets the filter level if not set already
//
// implements logger.FilterSetter
func (stream *MultiStream) SetFilterLevelIfUnset(level Level) {
	for _, s := range stream.streams {
		if setter, ok := s.(FilterSetter); ok {
			setter.SetFilterLevelIfUnset(level)
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

// Write writes the given Record
//
// implements logger.Streamer
func (stream *MultiStream) Write(record Record) error {
	errs := errors.MultiError{}
	for _, s := range stream.streams {
		if err := s.Write(record); err != nil {
			_ = errs.Append(errors.WithStack(err))
		}
	}
	return errs.AsError()
}

// ShouldWrite tells if the given level should be written to this stream
//
// implements logger.Streamer
func (stream *MultiStream) ShouldWrite(level Level) bool {
	return true
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
