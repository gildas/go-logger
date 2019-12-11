package logger

import (
	"time"
)

type BunyanConverter struct {
}

func (converter *BunyanConverter) Convert(record Record) Record {
	if value, ok := record["time"]; ok {
		if rtime, ok := value.(time.Time); ok {
			record["time"] = rtime.Format(time.RFC3339)
		}
	}
	return record
}