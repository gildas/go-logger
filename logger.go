package logger

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"time"
)

// Logger is a Logger that creates Bunyan's compatible logs (see: https://github.com/trentm/node-bunyan)
type Logger struct {
	stream    Streamer
	record    Record
	redactors []Redactor
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
	var (
		destinations = []string{}
		streams      = []Streamer{}
		records      = []Record{}
		redactors    = []Redactor{}
		filterLevel  = GetLevelFromEnvironment()
	)

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
		case Streamer:
			streams = append(streams, parameter)
		case Record:
			records = append(records, parameter)
		case Redactor:
			redactors = append(redactors, parameter)
		case *Redactor:
			redactors = append(redactors, *parameter)
		}
		// if param is a struct or pointer to struct, or interface
		// we should use it for the Topic, Scope
	}
	for _, destination := range destinations {
		streams = append(streams, CreateStreamWithDestination(destination))
	}
	logger := CreateWithStream(name, streams...)
	logger.SetFilterLevelIfUnset(filterLevel)
	for _, record := range records {
		for key, value := range record {
			logger.record.Set(key, value)
		}
	}
	logger.redactors = append(logger.redactors, redactors...)
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
		Set("tid", func() interface{} { return gettid() }).
		Set("topic", "main").
		Set("scope", "main").
		Set("v", 0)

	if len(streams) == 0 {
		return &Logger{CreateStreamWithDestination(), record, []Redactor{}}
	} else if len(streams) == 1 {
		return &Logger{streams[0], record, []Redactor{}}
	}
	return &Logger{&MultiStream{streams: streams}, record, []Redactor{}}
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
	return &Logger{log, NewRecord().Set(key, value), log.redactors}
}

// Recordf adds the given Record with formatted arguments
func (log *Logger) Recordf(key, value string, args ...interface{}) *Logger {
	return log.Record(key, fmt.Sprintf(value, args...))
}

// Records adds key, value pairs as Record objects
//
//   The key should be castable to a string.
//   If the last value is missing, its key is ignored
//
// E.g.: log.Records("key1", value1, "key2", value2)
func (log *Logger) Records(params ...interface{}) *Logger {
	if len(params) == 0 {
		return log
	}
	var key string
	record := NewRecord()
	for i, param := range params {
		if i%2 == 0 {
			key = param.(string)
		} else if len(key) > 0 {
			record.Set(key, param)
			key = ""
		}
	}
	return &Logger{log, record, log.redactors}
}

// Topic sets the Topic of this Logger
func (log *Logger) Topic(topic interface{}) *Logger {
	if topic == nil {
		topic = log.record["topic"]
	}
	return log.Record("topic", topic)
}

// Scope sets the Scope if this Logger
func (log *Logger) Scope(scope interface{}) *Logger {
	if scope == nil {
		scope = log.record["scope"]
	}
	return log.Record("scope", scope)
}

// Child creates a child Logger with a topic, a scope, and records
func (log *Logger) Child(topic, scope interface{}, params ...interface{}) *Logger {
	var key string
	if topic == nil {
		topic = log.record["topic"]
	}
	if scope == nil {
		scope = log.record["scope"]
	}
	record := NewRecord().Set("topic", topic).Set("scope", scope)
	newlog := &Logger{log, record, log.redactors}
	for _, param := range params {
		switch actual := param.(type) {
		case *Redactor:
			newlog.redactors = append(log.redactors, *actual)
		case Redactor:
			newlog.redactors = append(log.redactors, actual)
		case string:
			if len(key) == 0 {
				key = actual
			} else {
				record.Set(key, actual)
				key = ""
			}
		default:
			if len(key) > 0 {
				record.Set(key, actual)
				key = ""
			}
		}
	}
	return newlog
}

// GetRecord returns the Record field value for a given key
func (log *Logger) GetRecord(key string) interface{} {
	if value, found := log.record[key]; found {
		return value
	}
	if parent, ok := log.stream.(*Logger); ok {
		return parent.GetRecord(key)
	}
	return nil
}

// GetTopic returns the Record topic
func (log *Logger) GetTopic() string {
	return log.GetRecord("topic").(string)
}

// GetScope returns the Record scope
func (log *Logger) GetScope() string {
	return log.GetRecord("scope").(string)
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
//
// If the last argument is an error, a Record is added and the error string is added to the message
func (log *Logger) Errorf(msg string, args ...interface{}) {
	logWithErr := log

	if len(args) > 0 {
		errorInterface := reflect.TypeOf((*error)(nil)).Elem()
		last := args[len(args)-1]

		if last == nil {
			logWithErr.send(ERROR, msg, args[:len(args)-1]...)
			return
		} else if reflect.TypeOf(last).Implements(errorInterface) {
			logWithErr = log.Record("err", last)
			msg = msg + ", Error: %+v"
		}
	}
	logWithErr.send(ERROR, msg, args...)
}

// Fatalf traces a message at the FATAL Level
//
// If the last argument is an error, a Record is added and the error string is added to the message
func (log *Logger) Fatalf(msg string, args ...interface{}) {
	logWithErr := log

	if len(args) > 0 {
		errorInterface := reflect.TypeOf((*error)(nil)).Elem()
		last := args[len(args)-1]

		if last == nil {
			logWithErr.send(FATAL, msg, args[:len(args)-1]...)
			return
		} else if reflect.TypeOf(last).Implements(errorInterface) {
			logWithErr = log.Record("err", last)
			msg = msg + ", Error: %+v"
		}
	}
	logWithErr.send(FATAL, msg, args...)
}

// Memorylf traces memory usage at the given level and with the given message
func (log *Logger) Memorylf(level Level, msg string, args ...interface{}) {
	var mem runtime.MemStats

	runtime.ReadMemStats(&mem)
	if len(msg) > 0 {
		msg = msg + " Heap(Alloc = %s, System = %s), Stack(Alloc = %s, System = %s), NumGC = %d"
		args = append(
			args,
			bytesToString(mem.HeapAlloc),
			bytesToString(mem.Sys),
			bytesToString(mem.StackInuse),
			bytesToString(mem.StackSys),
			mem.NumGC,
		)
	} else {
		msg = "Heap(Alloc = %s, System = %s), Stack(Alloc = %s, System = %s), NumGC = %d"
		args = []interface{}{
			bytesToString(mem.HeapAlloc),
			bytesToString(mem.Sys),
			bytesToString(mem.StackInuse),
			bytesToString(mem.StackSys),
			mem.NumGC,
		}
	}
	log.send(level, msg, args...)
}

// Memoryf traces memory usage at the TRACE level with a given message
func (log *Logger) Memoryf(msg string, args ...interface{}) {
	log.Memorylf(TRACE, msg, args...)
}

// Memoryl traces memory usage at the given level
func (log *Logger) Memoryl(level Level) {
	log.Memorylf(level, "")
}

// Memory traces memory usage at the TRACE Level
func (log *Logger) Memory() {
	log.Memorylf(TRACE, "")
}

// send writes a message to the Sink
func (log *Logger) send(level Level, msg string, args ...interface{}) {
	if log.ShouldWrite(level, log.GetTopic(), log.GetScope()) {
		record := NewRecord()
		record["time"] = time.Now().UTC()
		record["level"] = level
		message := fmt.Sprintf(msg, args...)
		for _, redactor := range log.redactors {
			if msg, redacted := redactor.Redact(message); redacted {
				message = msg
				break
			}
		}
		record["msg"] = message
		if err := log.Write(record); err != nil {
			fmt.Fprintf(os.Stderr, "Logger error: %s\n", err)
		}
	}
}

func bytesToString(bytes uint64) string {
	if bytes >= 1024*1024*1024 {
		return fmt.Sprintf("%.2fGiB", float64(bytes)/1024.0/1024.0/1024.0)
	}
	if bytes >= 1024*1024 {
		return fmt.Sprintf("%.2fMiB", float64(bytes)/1024.0/1024.0)
	}
	if bytes >= 1024 {
		return fmt.Sprintf("%.2fKiB", float64(bytes)/1024.0)
	}
	return fmt.Sprintf("%dB", bytes)
}
