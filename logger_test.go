package logger_test

import (
	"fmt"
	"os"
	"testing"

	. "bitbucket.org/gildas_cherruel/go-logger"
	"github.com/stretchr/testify/assert"
)

type ErrorForTest struct {
  Errno string
  Code  int
}

func (e *ErrorForTest) Error() string {
        return fmt.Sprintf("Error %d - %s", e.Code, e.Errno)
}

func TestCreate(t *testing.T) {
	logger := Create("test")

	assert.NotNil(t, logger, "cannot create a logger")
	//assert.Equal(t, "main", logger.GetRecord("topic").(string))
}

func TestCreateWithDestination(t *testing.T) {
        logger := CreateWithDestination("test", "file:./log/test.log")

        if _, err := os.Stat("./log/test.log"); !os.IsNotExist(err) {
                err := os.Remove("./log/test.log")
                assert.Nil(t, err, "Could not remove the test log")
                _, err = os.Stat("./log/test.log");
                assert.True(t, os.IsNotExist(err), "The test log should not exist at this point")
        }
	assert.NotNil(t, logger, "cannot create a logger")
	logger.Infof("test of file destination")
        // TODO: We need some kind of Flush capability!
	//assert.Nil(t, logger.FlushSink(), "Could not flush the logger sink")
        //_, err := os.Stat("./log/test.log")
	//assert.False(t, os.IsNotExist(err), "The log was not created")
}

func TestAddRecord(t *testing.T) {
	logger := Create("test").Record("test", "test").(*Logger)

	assert.NotNil(t, logger, "cannot create a child logger")
	assert.Equal(t, "test", logger.GetRecord("test").(string))
}

func TestErrorWithDetails(t *testing.T) {
        logger := CreateWithDestination("test", "file:./log/test.log")
        err    := &ErrorForTest{ Errno: "ENOFOUND", Code: 12 }

	assert.NotNil(t, logger, "cannot create a logger")
        logger.Errorf("Got an error with number: %d", 2, err) 
}
