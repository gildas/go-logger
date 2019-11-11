package logger

// MultiStream is the Stream that writes to several streams
type MultiStream struct {
	streams []Stream
}

// Write writes the given Record
//   implements logger.Stream
func (stream *MultiStream) Write(record Record) error {
	for _, s := range stream.streams {
		if err := s.Write(record); err != nil {
			return err
		}
	}
	return nil
}

// ShouldWrite tells if the given level should be written to this stream
//   implements logger.Stream
func (stream *MultiStream) ShouldWrite(level Level) bool {
	for _, s := range stream.streams {
		if !s.ShouldWrite(level) {
			return false
		}
	}
	return true
}

// Flush flushes the stream (makes sure records are actually written)
//   implements logger.Stream
func (stream *MultiStream) Flush() {
}

// String gets a string version
//   implements the fmt.Stringer interface
func (stream MultiStream) String() string {
	return "MultiStream"
}