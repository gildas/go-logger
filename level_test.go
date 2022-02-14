package logger_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gildas/go-logger"
)

func TestCanParseLevel(t *testing.T) {
	assert.Equal(t, logger.INFO, logger.ParseLevel("INFO"), "Failed to parse level \"INFO\"")
	assert.Equal(t, logger.INFO, logger.ParseLevel("info"), "Failed to parse level \"INFO\"")
	assert.Equal(t, logger.UNSET, logger.ParseLevel("UNSET"), "Failed to parse level \"UNSET\"")
	assert.Equal(t, logger.NEVER, logger.ParseLevel("whatever"), "Failed to parse level \"whatever\" => NEVER")
}

func TestCanGetLevelFromEnvironment(t *testing.T) {
	os.Unsetenv("LOG_LEVEL")
	level := logger.GetLevelFromEnvironment()
	assert.Equal(t, logger.INFO, level, "Level should be INFO before being set in the environment")
	os.Setenv("DEBUG", "1")
	level = logger.GetLevelFromEnvironment()
	assert.Equal(t, logger.DEBUG, level, "Level should be INFO before being set in the environment and DEBUG is set")
	os.Setenv("LOG_LEVEL", "WARN")
	level = logger.GetLevelFromEnvironment()
	assert.Equal(t, logger.WARN, level, "Level should be WARN after being set in the environment")
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("DEBUG")
}

func TestCanStringLevel(t *testing.T) {
	assert.Equal(t, "INFO", logger.INFO.String())
	assert.Equal(t, "ERROR", logger.ERROR.String())
	bigLevel := logger.Level(220)
	assert.Equal(t, "ALWAYS", bigLevel.String())
	weirdLevel := logger.Level(47)
	assert.Equal(t, "WARN", weirdLevel.String())
	neverLevel := logger.UNSET
	assert.Equal(t, "UNSET", neverLevel.String())
}

func TestCanCompareLevels(t *testing.T) {
	levels := []logger.Level{logger.UNSET, logger.TRACE, logger.DEBUG, logger.INFO, logger.WARN, logger.ERROR, logger.FATAL, logger.ALWAYS}
	var filter logger.Level

	for index, filter := range levels[1:6] {
		for _, level := range levels[:index] {
			assert.Falsef(t, level.ShouldWrite(filter), "%s should not be writable with %s filter", level, filter)
		}
		for _, level := range levels[index+1:] {
			assert.Truef(t, level.ShouldWrite(filter), "%s should be writable with %s filter", level, filter)
		}
	}

	filter = logger.NEVER
	for _, level := range levels {
		assert.Falsef(t, level.ShouldWrite(filter), "%s should not be writable with %s filter", level, filter)
	}

	filter = logger.ALWAYS
	for _, level := range levels[:0] {
		assert.Falsef(t, level.ShouldWrite(filter), "%s should not be writable with %s filter", level, filter)
	}
	for _, level := range levels[1:] {
		assert.Truef(t, level.ShouldWrite(filter), "%s should be writable with %s filter", level, filter)
	}

	filter = logger.UNSET
	for _, level := range levels[:0] {
		assert.Falsef(t, level.ShouldWrite(filter), "%s should not be writable with %s filter", level, filter)
	}
	for _, level := range levels[1:] {
		assert.Truef(t, level.ShouldWrite(filter), "%s should be writable with %s filter", level, filter)
	}
}

func TestCanGetNextLevel(t *testing.T) {
	level := logger.DEBUG
	assert.Equal(t, logger.INFO, level.Next())
	level = logger.ALWAYS
	assert.Equal(t, logger.ALWAYS, level.Next())
}

func TestCanGetPreviousLevel(t *testing.T) {
	level := logger.DEBUG
	assert.Equal(t, logger.TRACE, level.Previous())
	level = logger.NEVER
	assert.Equal(t, logger.NEVER, level.Previous())
}