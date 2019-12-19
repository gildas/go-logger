package logger

import "time"

// PinoConverter is used to convert Records to be read by pino
type PinoConverter struct {
}

// Convert converts the Record into a Pino compatible Record
func (converter *PinoConverter) Convert(record Record) Record {
	if value, ok := record["level"]; ok {
		if level, ok := value.(Level); ok {
			record["level"] = int(level)
		}
	}
	if value, ok := record["time"]; ok {
		if rtime, ok := value.(time.Time); ok {
			record["time"] = rtime.UnixNano() / 1000000
		}
	}
	record["v"] = 1
	delete(record, "name")
	return record
}
