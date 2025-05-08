package logger

import (
	"strings"

	"github.com/gildas/go-core"
)

// Converter is used to convert a Record into another format
type Converter interface {
	Convert(record *Record) *Record
}

// GetConverterFromEnvironment fetches the Converter from the LOG_CONVERTER environment
func GetConverterFromEnvironment() Converter {
	return GetConverterFromEnvironmentWithPrefix("")
}

// GetConverterFromEnvironmentWithPrefix fetches the Converter from the LOG_CONVERTER environment with a prefix
func GetConverterFromEnvironmentWithPrefix(prefix EnvironmentPrefix) Converter {
	switch strings.ToLower(core.GetEnvAsString(string(prefix)+"LOG_CONVERTER", "bunyan")) {
	case "bunyan", "default":
		return &BunyanConverter{}
	case "stackdriver", "google", "gcp":
		return &StackDriverConverter{}
	case "aws", "cloudwatch":
		return &CloudWatchConverter{}
	default:
		return &BunyanConverter{}
	}
}
