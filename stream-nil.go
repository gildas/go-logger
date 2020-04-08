package logger

// NilStream is the Stream that writes nowhere
type NilStream struct {
}

// SetFilterLevel sets the filter level of all streams
func (stream *NilStream) SetFilterLevel(level Level) Streamer {
	return stream
}

// Write writes the given Record
func (stream *NilStream) Write(record Record) error {
	// implements logger.Stream
	// This stream does not write anything...
	return nil
}

// ShouldWrite tells if the given level should be written to this stream
func (stream *NilStream) ShouldWrite(level Level) bool {
	// implements logger.Stream
	return false
}

// Flush flushes the stream (makes sure records are actually written)
func (stream *NilStream) Flush() {
	// implements logger.Stream
}

// Close closes the stream
func (stream *NilStream) Close() {
}

// String gets a string version
func (stream NilStream) String() string {
	// implements the fmt.Stringer interface
	return "Stream to nil"
}
