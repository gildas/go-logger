package logger

import (
	"os"
	"strings"
)

type Level byte

const (
	NEVER Level = iota * 10
	TRACE
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
	ALWAYS = 255
)

// ParseLevel converts a string into a Level
func ParseLevel(value string) Level {
	if level, ok := map[string]Level{
	"NEVER":  NEVER,
	"TRACE":  TRACE,
	"DEBUG":  DEBUG,
	"INFO":   INFO,
	"WARN":   WARN,
	"ERROR":  ERROR,
	"FATAL":  FATAL,
	"ALWAYS": ALWAYS,
	}[strings.ToUpper(value)]; ok {
		return level
	}
	return NEVER
}

func GetLevelFromRecord(record Record) Level {
	if value, ok := record["level"]; ok {
		if level, ok := value.(Level); ok {
			return level
		}
	}
	return NEVER
}

// GetLevelFromEnvironment
func GetLevelFromEnvironment() Level {
	if value, ok := os.LookupEnv("LOG_LEVEL"); ok {
		return ParseLevel(value)
	}
	return NEVER
}

// ShouldWrite tells if the current level is writeable when compared to the given filter level
func (level Level) ShouldWrite(filter Level) bool {
	return filter == 0 || level >= filter
}

// String gets a string version
//   implements the fmt.Stringer interface
func (level Level) String() string {
	if level > ALWAYS {
		return "ALWAYS"
	}
	return []string{"NEVER", "TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL"}[level / 10]
}