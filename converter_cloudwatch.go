package logger

import "time"

// CloudWatchConverter is used to convert a Record for AWS CloudWatch
type CloudWatchConverter struct {
}

// Convert converts the Record into a StackDriver compatible Record
func (converter *CloudWatchConverter) Convert(record *Record) *Record {
	if value, found := record.Find("level"); found {
		if level, ok := value.(Level); ok {
			record.Data["severity"] = level.String()
		}
	}
	if value, ok := record.Find("time"); ok {
		if rtime, ok := value.(time.Time); ok {
			record.Data["time"] = rtime.Format(time.RFC3339)
		}
	}
	return record
}
