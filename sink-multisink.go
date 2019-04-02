package logger

import (
	"fmt"
	"github.com/chakrit/go-bunyan"
)

// MultiSink sends a Bunyan Record to several Sink objects
type MultiSink struct {
	sinks []bunyan.Sink
}

// NewMultiSink creates a MultiSink object
func NewMultiSink(sinks ...bunyan.Sink) *MultiSink {
	all := make([]bunyan.Sink, 0, len(sinks))
	for _, sink := range sinks {
		all = append(all, sink)
	}
	return &MultiSink{all}
}

func (sink MultiSink) Write(record bunyan.Record) error {
	for _, _sink := range sink.sinks {
		if err := _sink.Write(record); err != nil {
			return err
		}
	}
	return nil
}

func (sink MultiSink) String() string {
	return fmt.Sprintf("%v", sink.sinks)
}
