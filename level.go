package logger

import (
	"os"
	"strings"
)

// Level represents the log Level
// The higher the value the more chances to write in the Stream
type Level byte

const (
	// UNSET level means the level is not yet set
	UNSET Level = iota * 10
	// TRACE level should be used for entries that should be used by the developer of the app/package only
	TRACE
	// DEBUG level should be used for detailed logging as they tend to be noisy
	DEBUG
	// INFO level should be used as the standard level. Entries that really mean something to most people should go there
	INFO
	// WARN level should be used when the code considers a situation as not optimal but it can live with it
	WARN
	// ERROR level should be used when the code encounters an issue and normal flow is disrupted
	ERROR
	// FATAL level should be used when the code eoncounters an issue it cannot recover from
	FATAL
	// ALWAYS level should be used for entries that should always be logged, like app version, etc.
	ALWAYS Level = 255
	// NEVER level should be used for entries that should never be logged
	NEVER Level = 1
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

// GetLevelFromRecord retrieves the level from the given Record
func GetLevelFromRecord(record Record) Level {
	if value, ok := record["level"]; ok {
		if level, ok := value.(Level); ok {
			return level
		}
	}
	return NEVER
}

// GetLevelFromEnvironment retrieves the level from the environment LOG_LEVEL
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
func (level Level) String() string {
	// implements the fmt.Stringer interface
	if level > FATAL {
		return "ALWAYS"
	}
	if level == UNSET {
		return "UNSET"
	}
	return []string{"NEVER", "TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL"}[level/10]
}
