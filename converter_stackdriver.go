package logger

import (
	"cloud.google.com/go/logging"
)

type StackDriverConverter struct {
}

func (converter *StackDriverConverter) Convert(record Record) Record {
	// StackDriver special fields: https://cloud.google.com/logging/docs/agent/configuration#special-fields
	record["severity"] = severity(record["level"])
	record["message"] = record["msg"]
	delete(record, "level")
	delete(record, "name")
	delete(record, "msg")
	delete(record, "v")
	return record
}

func severity(value interface{}) logging.Severity {
	switch level := value.(type) {
	case Level:
		switch level {
		case TRACE:
			return logging.Debug
		case DEBUG:
			return logging.Debug
		case INFO:
			return logging.Info
		case WARN:
			return logging.Warning
		case ERROR:
			return logging.Error
		case FATAL:
			return logging.Critical
		case ALWAYS:
			return logging.Emergency
		default:
			return logging.Info
		}
	default:
		return logging.Info
	}
}