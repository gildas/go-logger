package logger

// NilStream is the Stream that writes nowhere
type NilStream struct {
}

// SetFilterLevel sets the filter level of all streams
//
// implements logger.FilterSetter
func (stream *NilStream) SetFilterLevel(level Level) {
}

// SetFilterLevelIfUnset sets the filter level if not set already
//
// implements logger.FilterSetter
func (stream *NilStream) SetFilterLevelIfUnset(level Level) {
}

// FilterMore tells the stream to filter more
//
// The stream will filter more if it is not already at the highest level.
// Which means less log messages will be written to the stream
//
// Example: if the stream is at DEBUG, it will be filtering at INFO
//
// implements logger.FilterModifier
func (stream *NilStream) FilterMore() {
}

// FilterLess tells the stream to filter less
//
// The stream will filter less if it is not already at the lowest level.
// Which means more log messages will be written to the stream
//
// Example: if the stream is at INFO, it will be filtering at DEBUG
//
// implements logger.FilterModifier
func (stream *NilStream) FilterLess() {
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
