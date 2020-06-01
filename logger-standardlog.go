package logger

import (
	"io"
	stdlog "log"
)

// Writer gets a n io.Writer to use with the standard log (for example)
func (log *Logger) Writer(level ...Level) io.Writer {
	if len(level) == 0 {
		level = []Level{INFO}
	}
	return &loggerWriter{logger: log, level: level[0]}
}

func (log *Logger) AsStandardLog(level ...Level) *stdlog.Logger {
	return stdlog.New(log.Writer(level...), "", 0)
}