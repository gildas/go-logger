package logger

type BunyanConverter struct {
}

func (converter *BunyanConverter) Convert(record Record) Record {
	return record
}