package logger

// Stream is the interface a Logger writes to
type Stream interface {
	Write(record Record) error
	ShouldWrite(level Level) bool
	Flush()
}