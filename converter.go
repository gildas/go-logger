package logger

type Converter interface {
	Convert(record Record) Record
}