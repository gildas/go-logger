package logger

// NilStream is the Stream that writes nowhere
type NilStream struct {
}

// SetFilterLevel sets the filter level of all streams
func (stream *NilStream) SetFilterLevel(level Level) Streamer {
	return stream
}

// Write writes the given Record
//   implements logger.Stream
func (stream *NilStream) Write(record Record) error {
	// This stream does not write anything...
	return nil
}

// ShouldWrite tells if the given level should be written to this stream
//   implements logger.Stream
func (stream *NilStream) ShouldWrite(level Level) bool {
	return false
}

// Flush flushes the stream (makes sure records are actually written)
//   implements logger.Stream
func (stream *NilStream) Flush() {
}

// String gets a string version
//   implements the fmt.Stringer interface
func (stream NilStream) String() string {
	return "Stream to nil"
}