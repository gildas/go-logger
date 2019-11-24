package logger_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/gorilla/mux"
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

func (suite *LoggerSuite) TestShouldPanicWithNoLogger() {
	defer func() {
		suite.Assert().NotNil(recover(), "logger.Must did not panic")
	}()
	_ = logger.Must(nil, nil)
}

func (suite *LoggerSuite) TestShouldPanicWithError() {
	defer func() {
		suite.Assert().NotNil(recover(), "logger.Must did not panic")
	}()
	_ = logger.Must(nil, &ErrorForTest{"error", 400})
}

func (suite *LoggerSuite) TestCanCreateLogWithMust() {
	log := logger.Must(logger.Create("test"), nil)
	suite.Require().NotNil(log, "cannot create a logger.Logger")
}

func (suite *LoggerSuite) TestCanCreateWithNil() {
	log := logger.CreateIfNil(nil, "test")
	suite.Require().NotNil(log, "cannot create a logger.Logger")
	log2 := logger.CreateIfNil(log, "test")
	suite.Require().NotNil(log2, "cannot create a logger.Logger")
}

func (suite *LoggerSuite) TestCanLoadAndSaveWithContext() {
	log := logger.Create("test")
	suite.Require().NotNil(log, "cannot create a logger.Logger")
	ctx := log.ToContext(context.Background())
	restored, err := logger.FromContext(ctx)
	suite.Assert().Nil(err, "Failed to retrieve a Logger from a context")
	suite.Assert().NotNil(restored, "cannot retrieve a logger.Logger from a context")
	_, err = logger.FromContext(context.Background())
	suite.Assert().NotNil(err, "Failed to retrieve a Logger from a context")
}

func (suite *LoggerSuite) TestCanAddRecord() {
	log := logger.Create("test")

	suite.Require().NotNil(log, "cannot create a logger.Logger")
	log = log.Record("test", "test")
	suite.Require().NotNil(log, "Failed to add a Record \"test\"")
	suite.Assert().Equal("test", log.GetRecord("test").(string))
	log = log.Recordf("testf", "now: %s", time.Now())
	suite.Assert().NotNil(log.GetRecord("testf"), "Failed to add Recordf")
	log2 := log.Records()
	suite.Assert().Equal(log, log2, "Failed to add no records")
	suite.Assert().Nil(log.GetRecord("unknown"), "Logger should not have a Record \"unknown\"")
}

func (suite *LoggerSuite) TestCanSetTopic() {
	log := logger.Create("test")
	suite.Require().NotNil(log, "cannot create a logger.Logger")
	log = log.Topic("topic_test")
	suite.Require().NotNil(log.GetRecord("topic"), "Failed to set a Topic")
	suite.Assert().Equal("topic_test", log.GetRecord("topic").(string), "Failed to set a Topic")
}

func (suite *LoggerSuite) TestCanSetScope() {
	log := logger.Create("test")
	suite.Require().NotNil(log, "cannot create a logger.Logger")
	log = log.Scope("scope_test")
	suite.Require().NotNil(log.GetRecord("scope"), "Failed to set a Scope")
	suite.Assert().Equal("scope_test", log.GetRecord("scope").(string), "Failed to set a Scope")
}

func (suite *LoggerSuite) TestCanLogAtTrace() {
	log, teardown := CreateLogger(suite.T(), "test.log", true)
	defer teardown()

	suite.Require().NotNil(log, "cannot create a logger.Logger")
	log.Tracef("test of file destination")
}

func (suite *LoggerSuite) TestCanLogAtDebug() {
	log, teardown := CreateLogger(suite.T(), "test.log", true)
	defer teardown()

	suite.Require().NotNil(log, "cannot create a logger.Logger")
	log.Debugf("test of file destination")
}

func (suite *LoggerSuite) TestCanLogAtInfo() {
	log, teardown := CreateLogger(suite.T(), "test.log", true)
	defer teardown()

	suite.Require().NotNil(log, "cannot create a logger.Logger")
	log.Infof("test of file destination")
}

func (suite *LoggerSuite) TestCanLogAtWarn() {
	log, teardown := CreateLogger(suite.T(), "test.log", true)
	defer teardown()

	suite.Require().NotNil(log, "cannot create a logger.Logger")
	log.Warnf("test of file destination")
}

func (suite *LoggerSuite) TestCanLogErrorWithDetails() {
	log, teardown := CreateLogger(suite.T(), "test.log", true)
	defer teardown()

	suite.Require().NotNil(log, "cannot create a logger.Logger")

	err := &ErrorForTest{Errno: "ENOFOUND", Code: 12}
	log.Errorf("Got an error with number: %d", 2, err)
}

func (suite *LoggerSuite) TestCanLogAtFatal() {
	log, teardown := CreateLogger(suite.T(), "test.log", true)
	defer teardown()

	suite.Require().NotNil(log, "cannot create a logger.Logger")
	err := &ErrorForTest{"testerror", 400}
	log.Fatalf("test of file destination", err)
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

func ExampleFailsLoggingWithBogusStream() {
	output := CaptureStderr(func() {
		log := logger.Create("test", &BogusStream{})

		log.Infof("test")
	})
	fmt.Println(output)
	// Output: Logger error: This Stream is Bogus
}

func FakeHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log, err := logger.FromContext(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		
		w.Write([]byte(fmt.Sprintf("%s", log)))
	})
}

func (suite *LoggerSuite) TestLoggerHttpHandler() {
	log := logger.Create("test")
	suite.Require().NotNil(log, "cannot create a logger.Logger")
	req, err := http.NewRequest("GET", "/", nil)
	suite.Require().Nil(err, "Failed to created an HTTP Request")

	rec := httptest.NewRecorder()
	router := mux.NewRouter()
	router.Methods("GET").Path("/").Handler(log.HttpHandler()(FakeHandler()))
	router.ServeHTTP(rec, req)
}