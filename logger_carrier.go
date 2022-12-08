package logger

// LoggerCarrier describes structs that carry a Logger
type LoggerCarrier interface {
	GetLogger() *Logger
}
