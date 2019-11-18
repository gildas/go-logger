package logger

// Write writes the given Record
//   implements logger.Stream
func (log *Logger) Write(record Record) error {
	record.Merge(log.record)
	return log.stream.Write(record)
}

// ShouldWrite tells if the given level should be written to this stream
//   implements logger.Stream
func (log *Logger) ShouldWrite(level Level) bool {
	return log.stream.ShouldWrite(level)
}

// Flush flushes the stream (makes sure records are actually written)
//   implements logger.Stream
func (log *Logger) Flush() {
	log.stream.Flush()
}