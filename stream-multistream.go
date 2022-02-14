package logger

import "github.com/gildas/go-errors"

// MultiStream is the Stream that writes to several streams
type MultiStream struct {
	streams []Streamer
}

// SetFilterLevel sets the filter level of all streams
//
// implements logger.Streamer
func (stream *MultiStream) SetFilterLevel(level Level) Streamer {
	for _, s := range stream.streams {
		s.SetFilterLevel(level)
	}
	return stream
}

// SetFilterLevelIfUnset sets the filter level if not set already
//
// implements logger.Streamer
func (stream *MultiStream) SetFilterLevelIfUnset(level Level) Streamer {
	for _, s := range stream.streams {
		s.SetFilterLevelIfUnset(level)
	}
	return stream
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
