package logger

// Streamer is the interface a Logger writes to
type Streamer interface {
	Write(record Record) error
	ShouldWrite(level Level) bool
	Flush()
}