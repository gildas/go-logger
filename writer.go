package logger

// loggerWriter is an io.Writer
type loggerWriter struct {
	level  Level
	logger *Logger
}

func (writer loggerWriter) Write(data []byte) (n int, err error) {
	writer.logger.send(writer.level, string(data))
	return len(data), nil
}