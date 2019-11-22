package logger_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gildas/go-logger"
)

func TestCanParseLevel(t *testing.T) {
	assert.Equal(t, logger.INFO,  logger.ParseLevel("INFO"))
	assert.Equal(t, logger.INFO,  logger.ParseLevel("info"))
	assert.Equal(t, logger.NEVER, logger.ParseLevel("whatever"))
}

func TestCanGetLevelFromEnvironment(t *testing.T) {
	os.Unsetenv("LOG_LEVEL")
	level := logger.GetLevelFromEnvironment()
	assert.Equal(t, logger.NEVER, level, "Level should be NEVER before being set in the environment")
	os.Setenv("LOG_LEVEL", "INFO")
	level = logger.GetLevelFromEnvironment()
	assert.Equal(t, logger.INFO, level, "Level should be INFO after being set in the environment")
	os.Unsetenv("LOG_LEVEL")
}

func TestCanStringLevel(t *testing.T) {
	assert.Equal(t, "INFO",  logger.INFO.String())
	assert.Equal(t, "ERROR", logger.ERROR.String())
	bigLevel := logger.Level(220)
	assert.Equal(t, "ALWAYS", bigLevel.String())
	weirdLevel := logger.Level(47)
	assert.Equal(t, "WARN", weirdLevel.String())
}

func TestCanCompareProperly(t *testing.T) {
	assert.True(t , logger.INFO.ShouldWrite(logger.DEBUG), "INFO should be writable with DEBUG filter")
	assert.False(t, logger.INFO.ShouldWrite(logger.WARN),  "INFO should not be writable with WARN filter")
}