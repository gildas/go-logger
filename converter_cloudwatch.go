package logger

import "time"

// CloudWatchConverter is used to convert a Record for AWS CloudWatch
type CloudWatchConverter struct {
}

// Convert converts the Record into a StackDriver compatible Record
func (converter *CloudWatchConverter) Convert(record Record) Record {
	if level, ok := record["level"].(Level); ok {
		record["severity"] = level.String()
	}
	if value, ok := record["time"]; ok {
		if rtime, ok := value.(time.Time); ok {
			record["time"] = rtime.Format(time.RFC3339)
		}
	}
	return record
}
