package logger_test

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/gildas/go-logger"
)

func TestCanParseLevel(t *testing.T) {
	assert.Equal(t, logger.INFO,  logger.ParseLevel("INFO"))
	assert.Equal(t, logger.INFO,  logger.ParseLevel("info"))
	assert.Equal(t, logger.NEVER, logger.ParseLevel("whatever"))
}

func TestCanStringLevel(t *testing.T) {
	assert.Equal(t, "INFO",  logger.INFO.String())
	assert.Equal(t, "ERROR", logger.ERROR.String())
}

func TestCanCompareProperly(t *testing.T) {
	assert.True(t , logger.INFO.ShouldWrite(logger.DEBUG), "INFO should be writable with DEBUG filter")
	assert.False(t, logger.INFO.ShouldWrite(logger.WARN),  "INFO should not be writable with WARN filter")
}