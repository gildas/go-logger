package logger

import (
	"time"
)

// BunyanConverter is the default Converter
type BunyanConverter struct {
}

// Convert converts the Record into a Bunyan compatible Record
func (converter *BunyanConverter) Convert(record *Record) *Record {
	if value, ok := record.Find("time"); ok {
		if rtime, ok := value.(time.Time); ok {
			record.Data["time"] = rtime.Format(time.RFC3339)
		}
	}
	return record
}
