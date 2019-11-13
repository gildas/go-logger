package logger_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gildas/go-logger"
)

type ErrorForTest struct {
	Errno string
	Code  int
}

func (e *ErrorForTest) Error() string {
	return fmt.Sprintf("Error %d - %s", e.Code, e.Errno)
}

func TestCanCreateSimple(t *testing.T) {
	log := logger.Create("test")

	require.NotNil(t, log, "cannot create a logger.Logger")
	assert.Equal(t, "main", log.GetRecord("topic").(string))
	//assert.IsType(t, logger.StdoutStream, log.stream, "The logger stream is not stdout")
}

func TestCanCreateWithDestination(t *testing.T) {
	log, teardown := CreateLogger(t, "test.log", false)
	defer teardown()

	require.NotNil(t, log, "cannot create a logger.Logger")
}

func TestCanAddRecord(t *testing.T) {
	log := logger.Create("test")

	require.NotNil(t, log, "cannot create a logger.Logger")
	log = log.Record("test", "test")
	require.NotNil(t, log, "Failed to add a logger.Logger")
	assert.Equal(t, "test", log.GetRecord("test").(string))
}

func TestCanLogAtInfo(t *testing.T) {
	log, teardown := CreateLogger(t, "test.log", true)
	defer teardown()

	require.NotNil(t, log, "cannot create a logger.Logger")
	log.Infof("test of file destination")
}

func TestCanLogErrorWithDetails(t *testing.T) {
	log, teardown := CreateLogger(t, "test.log", true)
	defer teardown()

	require.NotNil(t, log, "cannot create a logger.Logger")

	err := &ErrorForTest{Errno: "ENOFOUND", Code: 12}
	log.Errorf("Got an error with number: %d", 2, err)
}

func TestCanLogNested(t *testing.T) {
	log, teardown := CreateLogger(t, "test.log", true)
	defer teardown()

	require.NotNil(t, log, "cannot create a logger.Logger")
	log.Infof("test with main topic")
	{
		innerlog := log.Child("inner", "local", "temperature", "high")

		time.Sleep(1 * time.Second)
		innerlog.Infof("testing with inner topic")
		{
			innerMostLog := innerlog.Records("temperature", "low", "wind", "strong")

			innerMostLog.Debugf("testing with inner most log")
		}
	}
	log.Infof("test with main topic is over")
}

func TestCanLogWithFilter(t *testing.T) {
	folder, teardown := CreateTempDir(t)
	defer teardown()
	path := filepath.Join(folder, "test.log")
	stream := &logger.FileStream{Path: path, FilterLevel: logger.INFO, Unbuffered: true}
	log := logger.CreateWithStream("test", stream)

	log.Record("bello", "banana").Record("だれ", "Me").Infof("Log at INFO")
	log.Record("stuff", "other").Record("thing", "shiny").Debugf("Log at DEBUG")
	log.Flush()

	content, err := ioutil.ReadFile(stream.Path)
	require.Nil(t, err, "Failed to read %s", stream.Path)

	record := &logger.Record{}
	err = json.Unmarshal(content, &record)
	require.Nil(t, err, "Failed to unmarshal %s", stream.Path)
	//assert.Contains(t, record, "bello")
}