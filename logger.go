package logger

import (
	"fmt"
	"context"
	"errors"
	"html"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/satori/go.uuid"
	"github.com/chakrit/go-bunyan"
)

// Log is the object that allows to log stuff
type Log = bunyan.Log

// Logger is our Log implementation (based on bunyan.Logger)
type Logger struct {
	sink   bunyan.Sink
	record bunyan.Record
}

type key int

// ContextKey is the key for logger child stored in Context
const ContextKey key = iota

// Create creates a new Logger
func Create(name string) *Logger {
	destination, _ := os.LookupEnv("LOG_DESTINATION")
        return CreateWithDestination(name, destination)
}

// CreateWithDestination creates a new Logger sinking to the given destination
func CreateWithDestination(name, destination string) *Logger {
	var sink bunyan.Sink

	destination = strings.ToLower(destination)
	if "stackdriver" == destination {
		sink = NewMultiSink(bunyan.StdoutSink(), NewStackDriverSink())
	} else if "gcp" == destination {
		sink = NewGCPSink()
        } else if strings.HasPrefix(destination, "file:") {
                sink = bunyan.FileSink(strings.TrimPrefix(destination, "file:"))
        } else {
		sink = bunyan.StdoutSink()
	}

	logger:= &Logger{sink, bunyan.NewRecord()}
	
	return  logger.
			Record("name", name).
			Include(bunyan.LogVersionInfo(0)).
			Include(bunyan.PidInfo()).
			Include(ThreadIdInfo()).
			Include(bunyan.HostnameInfo()).
			Include(bunyan.TimeInfo()).
			Include(TopicInfo("main")).
			Include(ScopeInfo("main")).(*Logger)
}

// CreateWithSink creates a new Logger attacked to a given sink
func CreateWithSink(sink bunyan.Sink) *Logger {
	return &Logger{sink, bunyan.NewRecord()}
}

// Write writes the given records to the sink (intergafe bunyan.Log)
func (l *Logger) Write(record bunyan.Record) error {
	record.TemplateMerge(l.record)
	return l.sink.Write(record)
}

// Include returns a new Logger that records the Info
func (l *Logger) Include(info bunyan.Info) Log {
	return CreateWithSink(bunyan.InfoSink(l, info))
}

// Record adds the given Info to the Log
func (l *Logger) Record(key string, value interface{}) Log {
	builder := CreateWithSink(l)
	builder.record[key] = value
	return builder
}

// GetRecord returns the Record field value for a given key
func (l *Logger) GetRecord(key string) interface{} {
	value := l.record[key]

	if value != nil { return value }
	// TODO: find a way to traverse the sinks
	return nil
}

// Recordf adds the given Info with formatted arguments
func (l *Logger) Recordf(key, value string, args ...interface{}) Log {
	return l.Record(key, fmt.Sprintf(value, args...))
}

// Child creates a new child Logger
func (l *Logger) Child() Log {
	return CreateWithSink(l)
}

// Tracef traces a message at the TRACE Level
func (l *Logger) Tracef(msg string, args ...interface{}) { l.send(bunyan.TRACE, msg, args...) }

// Debugf traces a message at the DEBUG Level
func (l *Logger) Debugf(msg string, args ...interface{}) { l.send(bunyan.DEBUG, msg, args...) }

// Infof traces a message at the INFO Level
func (l *Logger) Infof(msg string, args ...interface{})  { l.send(bunyan.INFO,  msg, args...) }

// Warnf traces a message at the WARN Level
func (l *Logger) Warnf(msg string, args ...interface{})  { l.send(bunyan.WARN,  msg, args...) }

// Errorf traces a message at the ERROR Level
// If the last argument is an error, a Record is added and the error string is added to the message
func (l *Logger) Errorf(msg string, args ...interface{}) {
    log := l

    if len(args) > 0 {
        errorInterface := reflect.TypeOf((*error)(nil)).Elem()
        last := args[len(args) - 1]

        if reflect.TypeOf(last).Implements(errorInterface) {
            log = l.Record("err", last).(*Logger)
            msg = msg + ", Error: %s"
        }
    }
    log.send(bunyan.ERROR, msg, args...)
}

// Fatalf traces a message at the FATAL Level
func (l *Logger) Fatalf(msg string, args ...interface{}) {
    log := l

    if len(args) > 0 {
        errorInterface := reflect.TypeOf((*error)(nil)).Elem()
        last := args[len(args) - 1]

        if reflect.TypeOf(last).Implements(errorInterface) {
            log = l.Record("err", last).(*Logger)
            msg = msg + ", Error: %s"
        }
    }
    log.send(bunyan.FATAL, msg, args...)
}

// FromContext retrieves the Logger stored in the context
func FromContext(context context.Context) (*Logger, error) {
	logger := context.Value(ContextKey).(*Logger)
	if logger != nil {
		return logger, nil
	}
	return nil, errors.New("Context does not contain any Logger")
}

// ToContext stores the Logger in the given context
func (l *Logger) ToContext(parent context.Context) context.Context {
	return context.WithValue(parent, ContextKey, l)
}

// Handler function will wrap an http handler with extra logging information
func Handler(logger *Logger, h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Get a request identifier
		reqid := r.Header.Get("X-Line-Request-Id")
		if len(reqid) == 0 { reqid = r.Header.Get("X-Request-Id") }
		if len(reqid) == 0 { reqid = uuid.Must(uuid.NewV1()).String() }

		reqLogger := logger.Record("topic", "route").Record("scope", r.URL.Path).Record("reqid", reqid).Child()
		reqLogger.Infof("request start: %s %s", r.Method, html.EscapeString(r.URL.Path))
		h.ServeHTTP(w, r.WithContext(reqLogger.(*Logger).ToContext(r.Context())))
		duration := time.Now().Sub(start)
		reqLogger.Record("duration", duration.Seconds()).Infof("request finish: %s %s", r.Method, html.EscapeString(r.URL.Path))
	})
}

// send writes a message to the Sink
func (l *Logger) send(level bunyan.Level, msg string, args ...interface{}) {
	record := bunyan.NewRecord()
	record.SetMessagef(level, msg, args...)
	if err := l.Write(record); err != nil {
		fmt.Fprintf(os.Stderr, "Logger error: %s", err)
	}
}
