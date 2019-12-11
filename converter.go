package logger

import (
	"strings"
	"github.com/gildas/go-core"
)

type Converter interface {
	Convert(record Record) Record
}

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