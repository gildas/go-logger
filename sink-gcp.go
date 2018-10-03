package logger

import (
	"github.com/chakrit/go-bunyan"
)

// GCPSink sends Bunyan logs to Google Cloud (running insinde a Kubernetes Pod)
type GCPSink struct {
	_sink bunyan.Sink
}

// NewGCPSink creates a new GCPSink
func NewGCPSink() *GCPSink {
	return &GCPSink{bunyan.StdoutSink()}
}

func (sink GCPSink) Write(record bunyan.Record) error {
	record["severity"] = severity(record["level"])

	delete(record, "level")
	delete(record, "time")
	delete(record, "name")

	sink._sink.Write(record)
	return nil
}

func (sink GCPSink) String() string {
	return "GCPSink"
}
