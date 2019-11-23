package logger

import (
	"context"
	"github.com/pkg/errors"
)

type key int

// ContextKey is the key for logger child stored in Context
const ContextKey key = iota + 12583

// FromContext retrieves the Logger stored in the context
func FromContext(context context.Context) (*Logger, error) {
	if logger, ok := context.Value(ContextKey).(*Logger); ok {
		return logger, nil
	}
	return nil, errors.New("Context does not contain any Logger")
}

// ToContext stores the Logger in the given context
func (l *Logger) ToContext(parent context.Context) context.Context {
	return context.WithValue(parent, ContextKey, l)
}