package logger

// Write writes the given Record
func (log *Logger) Write(record Record) error {
	// implements logger.Stream
	record.Merge(log.record)
	return log.stream.Write(record)
}

// SetFilterLevel sets the filter level of the streamer
func (log *Logger) SetFilterLevel(level Level) Streamer {
	log.stream.SetFilterLevel(level)
	return log.stream
}

// ShouldWrite tells if the given level should be written to this stream
func (log *Logger) ShouldWrite(level Level) bool {
	// implements logger.Stream
	return log.stream.ShouldWrite(level)
}

// Flush flushes the stream (makes sure records are actually written)
func (log *Logger) Flush() {
	// implements logger.Stream
	log.stream.Flush()
}
