package logger

import (
	"fmt"
	"os"
	"reflect"
	"time"
)

// Logger is a Logger that creates Bunyan's compatible logs (see: https://github.com/trentm/node-bunyan)
type Logger struct {
	stream Streamer
	record Record
}

// Must gives the logger and panics if there is an error or if the Logger is nil
func Must(log *Logger, err error) *Logger {
	if err != nil {
		panic(err.Error())
	} else if log == nil {
		panic("Logger is nil")
	}
	return log
}

// Create2 is underwork...
func Create2(name string, parameters ...interface{}) *Logger {
	var (
		destination string
		origin      *Logger
		//sink        bunyan.Sink
	)
	for _, parameter := range parameters {
		if paramDestination, ok := parameter.(string); ok {
			destination = paramDestination
		}
		//if paramSink, ok := parameter.(bunyan.Sink); ok {
		//	sink = paramSink
		//}
		if paramLogger, ok := parameter.(*Logger); ok {
			origin = paramLogger
		}
		// if param is a struct or pointer to struct, or interface
		// we should use it for the Topic, Scope
	}
	if origin != nil {
		return origin
	}
	return CreateWithDestination(name, destination)
}

// Create creates a new Logger
func Create(name string) *Logger {
	return CreateWithDestination(name, "")
}

// CreateWithDestination creates a new Logger streaming to the given destination
func CreateWithDestination(name string, destination ...string) *Logger {
	return CreateWithStream(name, CreateStreamWithDestination(destination...))
}

func CreateWithStream(name string, stream Streamer) *Logger {
	hostname, _ := os.Hostname()
	record := NewRecord().
		Set("name", name).
		Set("hostname", hostname).
		Set("pid", os.Getpid()).
		Set("tid", func() interface{} { return Gettid() }).
		Set("time", func() interface{} { return time.Now().Format(time.RFC3339) }).
		Set("topic", "main").
		Set("scope", "main").
		Set("v", 0)

	if stream == nil {
		return &Logger{&NilStream{}, record}
	}
	return &Logger{stream, record}
}

// CreateIfNil creates a new Logger if the given Logger is nil, otherwise return the said Logger
func CreateIfNil(logger *Logger, name string) *Logger {
	if logger != nil {
		return logger
	}
	return CreateWithStream(name, nil)
}

// Record adds the given Record to the Log
func (log *Logger) Record(key string, value interface{}) *Logger {
	// This func requires Logger to be a Stream
	//   that allows us to nest Loggers
	return &Logger{log, NewRecord().Set(key, value)}
}

// Recordf adds the given Record with formatted arguments
func (log *Logger) Recordf(key, value string, args ...interface{}) *Logger {
	return log.Record(key, fmt.Sprintf(value, args...))
}

// Records adds key, value pairs as Record objects
// E.g.: log.Records("key1", value1, "key2", value2)
//   The key should be castable to a string
//   If the last value is missing, its key is ignored
func (log *Logger) Records(params ...interface{}) *Logger {
	var key string
	record := NewRecord()
	for i, param := range params {
		if i % 2 == 0 {
			key = param.(string)
		} else if len(key) > 0 {
			record.Set(key, param)
		}
	}
	return &Logger{log, record}
}

// Topic sets the Topic of this Logger
func (log *Logger) Topic(value interface{}) *Logger {
	return log.Record("topic", value)
}

// Scope sets the Scope if this Logger
func (log *Logger) Scope(value interface{}) *Logger {
	return log.Record("scope", value)
}

// Child creates a child Logger with a topic, a scope, and records
func (log *Logger) Child(topic, scope interface{}, params ...interface{}) *Logger {
	var key string
	record := NewRecord().Set("topic", topic).Set("scope", scope)
	for i, param := range params {
		if i % 2 == 0 {
			key = param.(string)
		} else if len(key) > 0 {
			record.Set(key, param)
		}
	}
	return &Logger{log, record}
}

// GetRecord returns the Record field value for a given key
func (log *Logger) GetRecord(key string) interface{} {
	value := log.record[key]

	if value != nil {
		return value
	}
	// TODO: find a way to traverse the parent Stream/Logger objects
	return nil
}

// String gets a string version
//   implements the fmt.Stringer interface
func (log Logger) String() string {
	return fmt.Sprintf("Logger(%s)", log.stream)
}

// Tracef traces a message at the TRACE Level
func (log *Logger) Tracef(msg string, args ...interface{}) { log.send(TRACE, msg, args...) }

// Debugf traces a message at the DEBUG Level
func (log *Logger) Debugf(msg string, args ...interface{}) { log.send(DEBUG, msg, args...) }

// Infof traces a message at the INFO Level
func (log *Logger) Infof(msg string, args ...interface{}) { log.send(INFO, msg, args...) }

// Warnf traces a message at the WARN Level
func (log *Logger) Warnf(msg string, args ...interface{}) { log.send(WARN, msg, args...) }

// Errorf traces a message at the ERROR Level
// If the last argument is an error, a Record is added and the error string is added to the message
func (log *Logger) Errorf(msg string, args ...interface{}) {
	logWithErr := log

	if len(args) > 0 {
		errorInterface := reflect.TypeOf((*error)(nil)).Elem()
		last := args[len(args)-1]

		if reflect.TypeOf(last).Implements(errorInterface) {
			logWithErr = log.Record("err", last)
			msg = msg + ", Error: %+v"
		}
	}
	logWithErr.send(ERROR, msg, args...)
}

// Fatalf traces a message at the FATAL Level
// If the last argument is an error, a Record is added and the error string is added to the message
func (log *Logger) Fatalf(msg string, args ...interface{}) {
	logWithErr := log

	if len(args) > 0 {
		errorInterface := reflect.TypeOf((*error)(nil)).Elem()
		last := args[len(args)-1]

		if reflect.TypeOf(last).Implements(errorInterface) {
			logWithErr = log.Record("err", last)
			msg = msg + ", Error: %+v"
		}
	}
	logWithErr.send(FATAL, msg, args...)
}

// send writes a message to the Sink
func (log *Logger) send(level Level, msg string, args ...interface{}) {
	if log.ShouldWrite(level) {
		record := NewRecord()
		record["level"] = level
		record["msg"]   = fmt.Sprintf(msg, args...)
		if err := log.Write(record); err != nil {
			fmt.Fprintf(os.Stderr, "Logger error: %s", err)
		}
	}
}