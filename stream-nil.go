package logger

// NilStream is the Stream that writes nowhere
type NilStream struct {
}

// SetFilterLevel sets the filter level of all streams
//
// implements logger.Streamer
func (stream *NilStream) SetFilterLevel(level Level) Streamer {
	return stream
}

// SetFilterLevelIfUnset sets the filter level if not set already
//
// implements logger.Streamer
func (stream *NilStream) SetFilterLevelIfUnset(level Level) Streamer {
	return stream
}

// Write writes the given Record
//
// implements logger.Streamer
func (stream *NilStream) Write(record Record) error {
	// This stream does not write anything...
	return nil
}

// ShouldWrite tells if the given level should be written to this stream
//
// implements logger.Streamer
func (stream *NilStream) ShouldWrite(level Level) bool {
	return false
}

// Flush flushes the stream (makes sure records are actually written)
//
// implements logger.Streamer
func (stream *NilStream) Flush() {
}

// Close closes the stream
//
// implements logger.Streamer
func (stream *NilStream) Close() {
}

// String gets a string version
//
// implements fmt.Stringer
func (stream NilStream) String() string {
	return "Stream to nil"
}
