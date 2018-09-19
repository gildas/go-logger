package logger_test

import (
	"testing"

	. "bitbucket.org/gildas_cherruel/go-logger"
	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	logger := Create("test")

	assert.NotNil(t, logger, "cannot create a logger")
	//assert.Equal(t, "main", logger.GetRecord("topic").(string))
}

func TestAddRecord(t *testing.T) {
	logger := Create("test").Record("test", "test").(*Logger)

	assert.NotNil(t, logger, "cannot create a child logger")
	assert.Equal(t, "test", logger.GetRecord("test").(string))
}
