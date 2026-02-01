package logger

import (
	"context"
	"github.com/gildas/go-errors"
)

type key int

// contextKey is the key for logger child stored in Context
const contextKey key = iota + 12583

// FromContext retrieves the Logger stored in the context
//
// Sources are either LoggerCarrier implemenations or Logger/*Logger objects.
//
// The first source that is a match is returned.
func FromContext(context context.Context, sources ...any) (*Logger, error) {
	if context != nil {
		if logger, ok := context.Value(contextKey).(*Logger); ok {
			return logger, nil
		}
	}
	for _, source := range sources {
		if logger, ok := source.(*Logger); ok {
			return logger, nil
		}
		if logger, ok := source.(Logger); ok {
			return &logger, nil
		}
		if carrier, ok := source.(LoggerCarrier); ok {
			return carrier.GetLogger(), nil
		}
	}
	return nil, errors.ArgumentMissing.With("Logger")
}

// ToContext stores the Logger in the given context
func (l *Logger) ToContext(parent context.Context) context.Context {
	return context.WithValue(parent, contextKey, l)
}
