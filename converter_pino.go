package logger

import "time"

type PinoConverter struct {
}

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
