package logger

import "github.com/gildas/go-errors"

// MultiStream is the Stream that writes to several streams
type MultiStream struct {
	streams []Streamer
}

// SetFilterLevel sets the filter level of all streams
func (stream *MultiStream) SetFilterLevel(level Level) Streamer {
	for _, s := range stream.streams {
		s.SetFilterLevel(level)
	}
	return stream
}

// Write writes the given Record
func (stream *MultiStream) Write(record Record) error {
	// implements logger.Stream
	errs := errors.MultiError{}
	for _, s := range stream.streams {
		if err := s.Write(record); err != nil {
			_ = errs.Append(errors.WithStack(err))
		}
	}
	return errs.AsError()
}

// ShouldWrite tells if the given level should be written to this stream
func (stream *MultiStream) ShouldWrite(level Level) bool {
	// implements logger.Stream
	return true
}

// Flush flushes the stream (makes sure records are actually written)
func (stream *MultiStream) Flush() {
	// implements logger.Stream
	for _, s := range stream.streams {
		s.Flush()
	}
}

// Close closes the stream
func (stream *MultiStream) Close() {
	for _, s := range stream.streams {
		s.Close()
	}
}

// String gets a string version
func (stream MultiStream) String() string {
	// implements the fmt.Stringer interface
	return "MultiStream"
}

// CreateMultiStream creates a MultiStream that contains all given Streamer objects
func CreateMultiStream(streams ...Streamer) Streamer {
	return &MultiStream{streams: streams}
}
