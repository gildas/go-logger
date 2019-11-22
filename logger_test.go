package logger_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/gildas/go-logger"
)

type LoggerSuite struct {
	suite.Suite
	Name string
}

type ErrorForTest struct {
	Errno string
	Code  int
}

func (e *ErrorForTest) Error() string {
	return fmt.Sprintf("Error %d - %s", e.Code, e.Errno)
}

func TestLoggerSuite(t *testing.T) {
	suite.Run(t, new(LoggerSuite))
}

func (suite *LoggerSuite) TestCanCreateSimple() {
	log := logger.Create("test")

	suite.Require().NotNil(log, "cannot create a logger.Logger")
	suite.Assert().Equal("main", log.GetRecord("topic").(string))
}

func (suite *LoggerSuite) TestCanCreateWithDestination() {
	log, teardown := CreateLogger(suite.T(), "test.log", false)
	defer teardown()

	suite.Require().NotNil(log, "cannot create a logger.Logger")
}

func (suite *LoggerSuite) TestCanAddRecord() {
	log := logger.Create("test")

	suite.Require().NotNil(log, "cannot create a logger.Logger")
	log = log.Record("test", "test")
	suite.Require().NotNil(log, "Failed to add a logger.Logger")
	suite.Assert().Equal("test", log.GetRecord("test").(string))
}

func (suite *LoggerSuite) TestCanLogAtInfo() {
	log, teardown := CreateLogger(suite.T(), "test.log", true)
	defer teardown()

	suite.Require().NotNil(log, "cannot create a logger.Logger")
	log.Infof("test of file destination")
}

func (suite *LoggerSuite) TestCanLogErrorWithDetails() {
	log, teardown := CreateLogger(suite.T(), "test.log", true)
	defer teardown()

	suite.Require().NotNil(log, "cannot create a logger.Logger")

	err := &ErrorForTest{Errno: "ENOFOUND", Code: 12}
	log.Errorf("Got an error with number: %d", 2, err)
}

func (suite *LoggerSuite) TestCanLogNested() {
	log, teardown := CreateLogger(suite.T(), "test.log", true)
	defer teardown()

	suite.Require().NotNil(log, "cannot create a logger.Logger")
	log.Infof("test with main topic")
	{
		expensiveLog := log.Record("key1", "value1").Record("key2", "value2")

		expensiveLog.Debugf("testing with expensive records")
		suite.T().Logf("Expensive Log (+2 Records): %s", expensiveLog)
	}
	{
		innerLog := log.Child("inner", "local", "temperature", "high", "vehicle", "car")

		time.Sleep(1 * time.Second)
		innerLog.Infof("testing with inner topic")
		suite.T().Logf("Inner Log (+3 Records): %s", innerLog)
		{
			innerMostLog := innerLog.Records("temperature", "low", "wind", "strong")

			innerMostLog.Debugf("testing with inner most log")
			suite.T().Logf("Innermost Log (+3 Records): %s", innerMostLog)
		}
	}
	log.Infof("test with main topic is over")
}

func (suite *LoggerSuite) TestCanLogWithFilter() {
	folder, teardown := CreateTempDir(suite.T())
	defer teardown()
	path := filepath.Join(folder, "test.log")
	stream := &logger.FileStream{Path: path, FilterLevel: logger.INFO, Unbuffered: true}
	log := logger.CreateWithStream("test", stream)

	log.Record("bello", "banana").Record("だれ", "Me").Infof("Log at INFO")
	log.Record("stuff", "other").Record("thing", "shiny").Debugf("Log at DEBUG")
	log.Flush()

	content, err := ioutil.ReadFile(stream.Path)
	suite.Require().Nil(err, "Failed to read %s", stream.Path)

	record := &logger.Record{}
	err = json.Unmarshal(content, &record)
	suite.Require().Nil(err, "Failed to unmarshal %s", stream.Path)
	//suite.Assert().Contains(record, "bello")
}