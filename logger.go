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

// Must returns the given logger or panics if there is an error or if the Logger is nil
func Must(log *Logger, err error) *Logger {
	if err != nil {
		panic(err.Error())
	} else if log == nil {
		panic("Logger is nil")
	}
	return log
}

// Create creates a new Logger
func Create(name string, parameters ...interface{}) *Logger {
	destinations := []string{}
	streams      := []Streamer{}
	records      := []Record{}
	filterLevel  := GetLevelFromEnvironment()

	for _, parameter := range parameters {
		switch parameter := parameter.(type) {
		case *Logger:
			if parameter != nil {
				return parameter
			}
		case string:
			destinations = append(destinations, parameter)
		case Level:
			filterLevel = parameter
		default:
			if streamer, ok := parameter.(Streamer); ok {
				streams = append(streams, streamer)
			} else if record, ok := parameter.(Record); ok {
				records = append(records, record)
			}
		}
		// if param is a struct or pointer to struct, or interface
		// we should use it for the Topic, Scope
	}
	for _, destination := range destinations {
		streams = append(streams, CreateStreamWithDestination(destination).SetFilterLevel(filterLevel))
	}
	logger := CreateWithStream(name, streams...)
	if len(records) > 0 {
		for _, record := range records {
			for key, value := range record {
				logger.record.Set(key, value)
			}
		}
	}
	return logger
}

// CreateWithDestination creates a new Logger streaming to the given destination(s)
func CreateWithDestination(name string, destinations ...string) *Logger {
	return CreateWithStream(name, CreateStreamWithDestination(destinations...))
}

// CreateWithStream creates a new Logger streaming to the given stream or list of streams
func CreateWithStream(name string, streams ...Streamer) *Logger {
	hostname, _ := os.Hostname()
	record := NewRecord().
		Set("name", name).
		Set("hostname", hostname).
		Set("pid", os.Getpid()).
		Set("tid", func() interface{} { return Gettid() }).
		Set("topic", "main").
		Set("scope", "main").
		Set("v", 0)

	if len(streams) == 0 {
		if value, ok := os.LookupEnv("DEBUG"); ok && value == "1" {
			return &Logger{&StdoutStream{Unbuffered: true}, record}
		} else {
			return &Logger{&StdoutStream{}, record}
		}
	} else if len(streams) == 1 {
		return &Logger{streams[0], record}
	}
	return &Logger{&MultiStream{streams: streams}, record}
}

// CreateIfNil creates a new Logger if the given Logger is nil, otherwise return the said Logger
func CreateIfNil(logger *Logger, name string) *Logger {
	if logger != nil {
		return logger
	}
	return CreateWithStream(name, &NilStream{})
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
	if len(params) == 0 {
		return log
	}
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
		record["time"]  = time.Now().Format(time.RFC3339)
		record["level"] = level
		record["msg"]   = fmt.Sprintf(msg, args...)
		if err := log.Write(record); err != nil {
			fmt.Fprintf(os.Stderr, "Logger error: %s\n", err)
		}
	}
}