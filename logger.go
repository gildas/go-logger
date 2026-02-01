package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gildas/go-errors"
)

// Logger is a Logger that creates Bunyan's compatible logs (see: https://github.com/trentm/node-bunyan)
type Logger struct {
	environmentPrefix EnvironmentPrefix
	stream            Streamer
	record            *Record
	redactors         []Redactor
}

// EnvironmentPrefix describes the prefix used for environment variables
type EnvironmentPrefix string

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
func Create(name string, parameters ...any) (logger *Logger) {
	var (
		destinations = []string{}
		streams      = []Streamer{}
		records      = []*Record{}
		redactors    = []Redactor{}
		filterLevels = ParseLevelsFromEnvironment()
		prefix       = EnvironmentPrefix("")
	)

	for _, parameter := range parameters {
		switch parameter := parameter.(type) {
		case *Logger:
			if parameter != nil {
				return parameter
			}
		case string:
			destinations = append(destinations, parameter)
		case EnvironmentPrefix:
			if len(parameter) > 0 {
				prefix = parameter
				filterLevels = ParseLevelsFromEnvironmentWithPrefix(prefix)
			}
		case Level:
			filterLevels.Set(parameter, "any", "any")
		case Streamer:
			streams = append(streams, parameter)
		case *Record:
			records = append(records, parameter)
		case Record:
			records = append(records, &parameter)
		case Redactor:
			redactors = append(redactors, parameter)
		case *Redactor:
			redactors = append(redactors, *parameter)
		}
		// if param is a struct or pointer to struct, or interface
		// we should use it for the Topic, Scope
	}

	hostname, _ := os.Hostname()
	record := NewRecord().
		Set("name", name).
		Set("hostname", hostname).
		Set("pid", os.Getpid()).
		Set("tid", func() any { return gettid() }).
		Set("topic", "main").
		Set("scope", "main").
		Set("v", 0)

	for _, destination := range destinations {
		streams = append(streams, CreateStreamWithPrefix(prefix, filterLevels, destination))
	}

	if len(streams) == 0 {
		logger = &Logger{prefix, CreateStreamWithPrefix(prefix, filterLevels), record, []Redactor{}}
	} else if len(streams) == 1 {
		logger = &Logger{prefix, streams[0], record, []Redactor{}}
	} else {
		logger = &Logger{prefix, &MultiStream{streams: streams}, record, []Redactor{}}
	}

	for _, record := range records {
		for key, value := range record.Data {
			logger.record.Set(key, value)
		}
	}
	logger.redactors = append(logger.redactors, redactors...)
	return logger
}

// CreateIfNil creates a new Logger if the given Logger is nil, otherwise return the said Logger
func CreateIfNil(logger *Logger, name string) *Logger {
	if logger != nil {
		return logger
	}
	return Create(name, &NilStream{})
}

// AddDestinations adds destinations to the Logger
func (log *Logger) AddDestinations(destinations ...any) {
	streams := []Streamer{}

	for _, raw := range destinations {
		switch destination := raw.(type) {
		case string:
			streams = append(streams, CreateStreamWithPrefix(log.environmentPrefix, log.GetFilterLevels(), destination))
		case Streamer:
			streams = append(streams, destination)
		}
	}

	if len(streams) > 0 {
		if multi, ok := log.stream.(*MultiStream); ok {
			multi.streams = append(multi.streams, streams...)
		} else {
			log.stream = &MultiStream{streams: append([]Streamer{log.stream}, streams...)}
		}
	}
}

// ResetDestinations resets the destinations to the Logger
//
// If no destinations are given, nothing happens
func (log *Logger) ResetDestinations(destinations ...any) {
	streams := []Streamer{}

	for _, raw := range destinations {
		switch destination := raw.(type) {
		case string:
			streams = append(streams, CreateStreamWithPrefix(log.environmentPrefix, log.GetFilterLevels(), destination))
		case Streamer:
			streams = append(streams, destination)
		}
	}

	if len(streams) > 1 {
		log.stream = &MultiStream{streams: streams}
	} else if len(streams) == 1 {
		log.stream = streams[0]
	}
}

// Record adds the given Record to the Log
func (log *Logger) Record(key string, value any) *Logger {
	// This func requires Logger to be a Stream
	//   that allows us to nest Loggers
	return &Logger{log.environmentPrefix, log, NewRecord().Set(key, value), log.redactors}
}

// RecordWithKeysToRedact adds the given Record to the Log
func (log *Logger) RecordWithKeysToRedact(key string, value any, keyToRedact ...string) *Logger {
	// This func requires Logger to be a Stream
	//   that allows us to nest Loggers
	return &Logger{log.environmentPrefix, log, NewRecord().Set(key, value).AddKeysToRedact(keyToRedact...), log.redactors}
}

// Recordf adds the given Record with formatted arguments
func (log *Logger) Recordf(key, value string, args ...any) *Logger {
	return log.Record(key, fmt.Sprintf(value, args...))
}

// Records adds key, value pairs as Record objects
//
//	The key should be castable to a string.
//	If the last value is missing, its key is ignored
//
// E.g.: log.Records("key1", value1, "key2", value2)
func (log *Logger) Records(params ...any) *Logger {
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
	return &Logger{log.environmentPrefix, log, record, log.redactors}
}

// RecordMap adds the given map as Record objects and returns a new Logger
//
// if the map is empty, the Logger is returned unchanged
func (log *Logger) RecordMap(values map[string]any) *Logger {
	if len(values) == 0 {
		return log
	}
	record := NewRecord()
	for key, value := range values {
		record.Set(key, value)
	}
	return &Logger{log.environmentPrefix, log, record, log.redactors}
}

// Topic sets the Topic of this Logger
func (log *Logger) Topic(topic any) *Logger {
	if topic == nil {
		topic = log.record.Get("topic")
	}
	return log.Record("topic", topic)
}

// Scope sets the Scope if this Logger
func (log *Logger) Scope(scope any) *Logger {
	if scope == nil {
		scope = log.record.Get("scope")
	}
	return log.Record("scope", scope)
}

// Child creates a child Logger with a topic, a scope, and records
func (log *Logger) Child(topic, scope any, params ...any) *Logger {
	var key string
	if topic == nil {
		topic = log.record.Get("topic")
	}
	if scope == nil {
		scope = log.record.Get("scope")
	}
	record := NewRecord().Set("topic", topic).Set("scope", scope)
	newlog := &Logger{log.environmentPrefix, log, record, log.redactors}
	for _, param := range params {
		switch actual := param.(type) {
		case *Redactor:
			newlog.redactors = append(newlog.redactors, *actual)
		case Redactor:
			newlog.redactors = append(newlog.redactors, actual)
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
func (log *Logger) GetRecord(key string) any {
	if value, found := log.record.Find(key); found {
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
func (log *Logger) Tracef(msg string, args ...any) { log.send(TRACE, msg, args...) }

// Debugf traces a message at the DEBUG Level
func (log *Logger) Debugf(msg string, args ...any) { log.send(DEBUG, msg, args...) }

// Infof traces a message at the INFO Level
func (log *Logger) Infof(msg string, args ...any) { log.send(INFO, msg, args...) }

// Warnf traces a message at the WARN Level
func (log *Logger) Warnf(msg string, args ...any) { log.send(WARN, msg, args...) }

// Errorf traces a message at the ERROR Level
//
// If the last argument is an error, a Record is added and the error string is added to the message
func (log *Logger) Errorf(msg string, args ...any) {
	logWithErr := log

	if len(args) > 0 {
		last := args[len(args)-1]

		if last == nil {
			logWithErr.send(ERROR, msg, args[:len(args)-1]...)
			return
		} else if err, ok := last.(error); ok {
			logWithErr = log.Record("err", err)
			msg = msg + ", Error: %+v"
		}
	}
	logWithErr.send(ERROR, msg, args...)
}

// Fatalf traces a message at the FATAL Level
//
// If the last argument is an error, a Record is added and the error string is added to the message
func (log *Logger) Fatalf(msg string, args ...any) {
	logWithErr := log

	if len(args) > 0 {
		last := args[len(args)-1]

		if last == nil {
			logWithErr.send(FATAL, msg, args[:len(args)-1]...)
			return
		} else if err, ok := last.(error); ok {
			logWithErr = log.Record("err", err)
			msg = msg + ", Error: %+v"
		}
	}
	logWithErr.send(FATAL, msg, args...)
}

// Memorylf traces memory usage at the given level and with the given message
func (log *Logger) Memorylf(level Level, msg string, args ...any) {
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
		args = []any{
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
func (log *Logger) Memoryf(msg string, args ...any) {
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

// send writes a message to the Stream
func (log *Logger) send(level Level, msg string, args ...any) {
	if log.ShouldWrite(level, log.GetTopic(), log.GetScope()) {
		record, release := NewPooledRecord()
		defer release()
		record.Set("time", time.Now().UTC())
		record.Set("level", level)
		if log.stream.ShouldLogSourceInfo() {
			if counter, file, line, ok := runtime.Caller(2); ok {
				funcname := runtime.FuncForPC(counter).Name()
				i := strings.LastIndex(funcname, "/")
				if i == -1 {
					i = 0 // main func typically has no slash
				}
				i += strings.Index(funcname[i:], ".")

				record.Set("file", filepath.Base(file))
				record.Set("line", line)
				record.Set("func", funcname[i+1:])
				record.Set("package", funcname[:i])
			}
		}
		message := fmt.Sprintf(msg, args...)
		for _, redactor := range log.redactors {
			if msg, redacted := redactor.Redact(message); redacted {
				message = msg
				break
			}
		}
		record.Set("msg", message)
		if err := log.Write(record); err != nil {
			fmt.Fprintf(os.Stderr, "Logger error: %+v\n", errors.RuntimeError.Wrap(err))
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
