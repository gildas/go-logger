package logger

import (
	"time"

	"cloud.google.com/go/logging"
)

// StackDriverConverter is used to convert a Record for StackDriver
type StackDriverConverter struct {
}

// Convert converts the Record into a StackDriver compatible Record
func (converter *StackDriverConverter) Convert(record Record) Record {
	// StackDriver special fields: https://cloud.google.com/logging/docs/agent/configuration#special-fields
	record["severity"] = converter.severity(record["level"])
	record["message"] = record["msg"]
	if value, ok := record["time"]; ok {
		if rtime, ok := value.(time.Time); ok {
			record["time"] = rtime.Format(time.RFC3339)
		}
	}
	return record
}

func (converter StackDriverConverter) severity(value interface{}) logging.Severity {
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
