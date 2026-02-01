package logger

import "time"

// PinoConverter is used to convert Records to be read by pino
type PinoConverter struct {
}

// Convert converts the Record into a Pino compatible Record
func (converter *PinoConverter) Convert(record *Record) *Record {
	if value, found := record.Find("level"); found {
		if level, ok := value.(Level); ok {
			record.Data["level"] = int(level)
		}
	}
	if value, found := record.Find("time"); found {
		if rtime, ok := value.(time.Time); ok {
			record.Data["time"] = rtime.UnixNano() / 1000000
		}
	}
	record.Data["v"] = 1
	record.Delete("name")
	return record
}
