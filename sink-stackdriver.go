package logger

// Check this: https://cloud.google.com/kubernetes-engine/docs/tutorials/authenticating-to-cloud-platform

import (
	"context"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/logging"
	"github.com/chakrit/go-bunyan"
)

// StackDriverSink sends Bunyan logs to Google StackDriver
type StackDriverSink struct {
	ctx    context.Context
	client *logging.Client
	target *logging.Logger
}

// NewStackDriverSink creates a new StackDriverSink
func NewStackDriverSink() *StackDriverSink {
	ctx := context.Background()
	if projectID, ok := os.LookupEnv("PROJECT_ID"); ok {
		if client, err := logging.NewClient(ctx, projectID); err != nil {
			fmt.Fprintf(os.Stderr, "Failure while creating StackDriver Sink to Project %s: %s\n", projectID, err)
		} else {
	//defer client.Close()
			target := client.Logger("whoareyou")
			return &StackDriverSink{ctx, client, target}
		}
	}
	return nil
}

// Write sends a Bunyan Record to a StackDriverSink
func (sink StackDriverSink) Write(record bunyan.Record) error {
	var stamp time.Time
	var err error

	severity := severity(record["level"])
	if stamp, err = time.Parse(time.RFC3339, record["time"].(string)); err != nil {
		stamp = time.Now()
	}
	delete(record, "level")
	delete(record, "time")
	delete(record, "name")
	sink.target.Log(logging.Entry{
		Timestamp: stamp,
		Severity:  severity,
		Payload:   record,
	})
	return nil
}

func severity(level interface{}) logging.Severity {
	switch level.(bunyan.Level) {
	case bunyan.DEBUG:
		return logging.Debug
	case bunyan.INFO:
		return logging.Info
	case bunyan.WARN:
		return logging.Warning
	case bunyan.ERROR:
		return logging.Error
	case bunyan.FATAL:
		return logging.Critical
	case bunyan.EVERYTHING:
		return logging.Emergency
	default:
		return logging.Info
	}
}

func (sink StackDriverSink) String() string {
	return "StackDriverSink"
}
