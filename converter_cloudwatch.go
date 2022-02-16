package logger

import "time"

// CloudWatchConverter is used to convert a Record for AWS CloudWatch
type CloudWatchConverter struct {
}

// Convert converts the Record into a StackDriver compatible Record
func (converter *CloudWatchConverter) Convert(record Record) Record {
	if value, ok := record["level"]; ok {
		if level, ok := value.(Level); ok {
			record["level"] = level.String()
		}
	} else {
		record["level"] = INFO.String()
	}
	record["message"] = record["msg"]
	if value, ok := record["time"]; ok {
		if rtime, ok := value.(time.Time); ok {
			record["time"] = rtime.Format(time.RFC3339)
		}
	}
	delete(record, "name")
	delete(record, "msg")
	delete(record, "v")
	return record
}
