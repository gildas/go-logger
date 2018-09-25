package logger_test

import (
	"path/filepath"
	"fmt"
	"io/ioutil"
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

func testCreateTempDir(t *testing.T) (string, func()) {
        dir, err := ioutil.TempDir("", "go_logger")
        if err != nil {
                t.Fatalf("Unable to create a temp folder for log files. Error: %s\n", err)
        }
        t.Logf("Log Temp Folder: %s", dir)
        return dir, func() { os.RemoveAll(dir) }
}

func TestCreate(t *testing.T) {
	logger := Create("test")

	assert.NotNil(t, logger, "cannot create a logger")
	//assert.Equal(t, "main", logger.GetRecord("topic").(string))
}

func TestCreateWithDestination(t *testing.T) {
        dir, teardown := testCreateTempDir(t); defer teardown()
        logger := CreateWithDestination("test", "file:" + filepath.Join(dir, "test.log"))

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
        dir, teardown := testCreateTempDir(t); defer teardown()
        logger := CreateWithDestination("test", "file:" + filepath.Join(dir, "test.log"))
        err    := &ErrorForTest{ Errno: "ENOFOUND", Code: 12 }

	assert.NotNil(t, logger, "cannot create a logger")
        logger.Errorf("Got an error with number: %d", 2, err) 
}

func TestFatalWithDetails(t *testing.T) {
        dir, teardown := testCreateTempDir(t); defer teardown()
        logger := CreateWithDestination("test", "file:" + filepath.Join(dir, "test.log"))
        err    := &ErrorForTest{ Errno: "ENOFOUND", Code: 12 }

	assert.NotNil(t, logger, "cannot create a logger")
        logger.Fatalf("Got an error with number: %d", 2, err)
}
