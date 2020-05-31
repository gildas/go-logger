package logger

// loggerWriter is an io.Writer
type loggerWriter struct {
	logger *Logger
}

func (writer loggerWriter) Write(data []byte) (n int, err error) {
	writer.logger.Infof(string(data))
	return len(data), nil
}