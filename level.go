package logger

import (
	"os"
	"strings"
)

type Level byte

const (
	UNSET Level = iota * 10
	TRACE
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
	ALWAYS Level = 255
	NEVER  Level = 1
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
	"UNSET":  UNSET,
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
	if value, ok := os.LookupEnv("DEBUG"); ok && value == "1" {
		return DEBUG
	}
	return INFO
}

// ShouldWrite tells if the current level is writeable when compared to the given filter level
func (level Level) ShouldWrite(filter Level) bool {
	if level == NEVER || level == UNSET {
		return false
	}
	return filter == ALWAYS || filter == UNSET || (filter != NEVER && level >= filter)
}

// String gets a string version
//   implements the fmt.Stringer interface
func (level Level) String() string {
	if level > FATAL {
		return "ALWAYS"
	}
	if level == UNSET {
		return "UNSET"
	}
	return []string{"NEVER", "TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL"}[level / 10]
}