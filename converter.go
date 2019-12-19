package logger

import (
	"strings"
	"github.com/gildas/go-core"
)

// Converter is used to convert a Record into another format
type Converter interface {
	Convert(record Record) Record
}

// GetConverterFromEnvironment fetches the Converter from the environment
func GetConverterFromEnvironment() Converter {
	switch strings.ToLower(core.GetEnvAsString("LOG_CONVERTER", "bunyan")) {
	case "bunyan", "default":
		return &BunyanConverter{}
	case "stackdriver", "google", "gcp":
		return &StackDriverConverter{}
	default:
		return &BunyanConverter{}
	}
}