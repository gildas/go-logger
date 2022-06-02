package logger_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/suite"

	"github.com/gildas/go-core"
	"github.com/gildas/go-errors"
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

func (suite *LoggerSuite) SetupSuite() {
	suite.Name = strings.TrimSuffix(reflect.TypeOf(*suite).Name(), "Suite")
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

func (suite *LoggerSuite) TestShouldFailLoadingFromContextWithoutLogger() {
	_, err := logger.FromContext(context.Background())
	suite.Require().NotNil(err, "Context should not contain a Logger")
	suite.Assert().True(errors.Is(err, errors.ArgumentMissing), "error should be an Argument Missing error")
	details := errors.ArgumentMissing.Clone()
	suite.Require().ErrorAs(err, &details, "Error chain should contain an errors.ArgumentMissing")
	suite.Assert().Equal("Logger", details.What, "Error's What is wrong")
}

func (suite *LoggerSuite) TestShouldFailLoadingFromNilContext() {
	_, err := logger.FromContext(nil) //nolint
	suite.Require().NotNil(err, "Context should not contain a Logger")
	suite.Assert().True(errors.Is(err, errors.ArgumentMissing), "error should be an Argument Missing error")
	details := errors.ArgumentMissing.Clone()
	suite.Require().ErrorAs(err, &details, "Error chain should contain an errors.ArgumentMissing")
	suite.Assert().Equal("Logger", details.What, "Error's What is wrong")
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
	suite.Require().NotNil(log.GetTopic(), "Failed to set a Topic")
	suite.Assert().Equal("topic_test", log.GetTopic(), "Failed to set a Topic")
}

func (suite *LoggerSuite) TestCanSetTopicWithNilValue() {
	log := logger.Create("test")
	suite.Require().NotNil(log, "cannot create a logger.Logger")
	log = log.Topic(nil)
	suite.Require().NotNil(log.GetTopic(), "Failed to set a Topic")
	suite.Assert().Equal("main", log.GetTopic(), "Failed to set a Topic")
}

func (suite *LoggerSuite) TestCanSetScope() {
	log := logger.Create("test")
	suite.Require().NotNil(log, "cannot create a logger.Logger")
	log = log.Scope("scope_test")
	suite.Require().NotNil(log.GetScope(), "Failed to set a Scope")
	suite.Assert().Equal("scope_test", log.GetScope(), "Failed to set a Scope")
}

func (suite *LoggerSuite) TestCanSetScopeWithNilValue() {
	log := logger.Create("test")
	suite.Require().NotNil(log, "cannot create a logger.Logger")
	log = log.Scope(nil)
	suite.Require().NotNil(log.GetScope(), "Failed to set a Scope")
	suite.Assert().Equal("main", log.GetScope(), "Failed to set a Scope")
}

func (suite *LoggerSuite) TestCanGetTopic() {
	log := logger.Create("test", &logger.StdoutStream{Unbuffered: true})
	suite.Assert().Equal("main", log.GetTopic())
}

func (suite *LoggerSuite) TestCanGetTopicFromChild() {
	log := logger.Create("test", &logger.StdoutStream{Unbuffered: true})
	suite.Assert().Equal("main", log.GetTopic())
	log = log.Child("child", "scope")
	suite.Assert().Equal("child", log.GetTopic())
}

func (suite *LoggerSuite) TestCanGetTopicFromChildWithAddedRecord() {
	log := logger.Create("test", &logger.StdoutStream{Unbuffered: true})
	suite.Assert().Equal("main", log.GetTopic())
	log = log.Child("child", "scope")
	suite.Assert().Equal("child", log.GetTopic())
	log = log.Record("data", "data")
	suite.Assert().Equal("child", log.GetTopic())
}

func (suite *LoggerSuite) TestCanGetTopicInheritedByChild() {
	log := logger.Create("test", &logger.StdoutStream{Unbuffered: true})
	suite.Assert().Equal("main", log.GetTopic())
	log = log.Child(nil, "scope")
	suite.Assert().Equal("main", log.GetTopic())
}

func (suite *LoggerSuite) TestCanSetLevelPerTopic() {
	log := logger.Create("test", &logger.StdoutStream{Unbuffered: true})

	suite.Assert().Truef(log.ShouldWrite(logger.INFO, "main", ""), "Logger should write INFO messages for main topic before it is configured")
	suite.Assert().Falsef(log.ShouldWrite(logger.DEBUG, "main", ""), "Logger should not write DEBUG messages for main topic before it is configured")

	log.SetFilterLevelForTopic(logger.DEBUG, "main")

	suite.Assert().Truef(log.ShouldWrite(logger.WARN, "", ""), "Logger should write WARN messages")
	suite.Assert().Falsef(log.ShouldWrite(logger.DEBUG, "", ""), "Logger should not write DEBUG messages")

	suite.Assert().Truef(log.ShouldWrite(logger.DEBUG, "main", ""), "Logger should write DEBUG messages for main topic")
	suite.Assert().Falsef(log.ShouldWrite(logger.TRACE, "main", ""), "Logger should not write TRACE messages for main topic")

	suite.Assert().Truef(log.ShouldWrite(logger.INFO, "another_topic", ""), "Logger should write INFO messages for another_topic topic")
	suite.Assert().Falsef(log.ShouldWrite(logger.DEBUG, "another_topic", ""), "Logger should not write DEBUG messages for another_topic topic")
}

func (suite *LoggerSuite) TestCanSetLevelPerTopicAndScope() {
	log := logger.Create("test", &logger.StdoutStream{Unbuffered: true})

	suite.Assert().Truef(log.ShouldWrite(logger.INFO, "main", "any"), "Logger should write INFO messages for main topic and any scope before it is configured")
	suite.Assert().Falsef(log.ShouldWrite(logger.DEBUG, "main", "any"), "Logger should not write DEBUG messages for main topic and any scope before it is configured")

	log.SetFilterLevelForTopicAndScope(logger.TRACE, "main", "specific")
	log.SetFilterLevelForTopic(logger.DEBUG, "main")

	suite.Assert().Truef(log.ShouldWrite(logger.DEBUG, "main", "any"), "Logger should write DEBUG messages for main topic and any scope")
	suite.Assert().Truef(log.ShouldWrite(logger.TRACE, "main", "specific"), "Logger should write TRACE messages for main topic and specific scope")
	suite.Assert().Falsef(log.ShouldWrite(logger.TRACE, "main", "any"), "Logger should not write TRACE messages for main topic and any scope")

	suite.Assert().Falsef(log.ShouldWrite(logger.DEBUG, "another_topic", "any"), "Logger should not write DEBUG messages for another_topic topic and any scope")
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

func ExampleLogger_Write_failsWithBogusStream() {
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

		count, err := w.Write([]byte(log.String()))
		if err != nil {
			log.Errorf("Failed to write response", err)
			return
		}
		log.Infof("Written %d bytes", count)
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

func (suite *LoggerSuite) TestCanUseWithIOWriter() {
	output := CaptureStdout(func() {
		logger := logger.Create("test", &logger.StdoutStream{Unbuffered: true})
		log := log.New(logger.Writer(), "", 0)
		log.Print("This is a Standard Log message")
	})
	pattern := regexp.MustCompile(`{"hostname":"[a-zA-Z_0-9\-\.]+","level":30,"msg":"This is a Standard Log message\\n","name":"test","pid":[0-9]+,"scope":"main","tid":[0-9]+,"time":"[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+Z","topic":"main","v":0}`)
	suite.Assert().Truef(pattern.MatchString(output), "Output is malformed: %s", output)
}

func (suite *LoggerSuite) TestCanUseWithIOWriterWithLevel() {
	output := CaptureStdout(func() {
		l := logger.Create("test", &logger.StdoutStream{Unbuffered: true})
		log := log.New(l.Writer(logger.WARN), "", 0)
		log.Print("This is a Standard Log message")
	})
	pattern := regexp.MustCompile(`{"hostname":"[a-zA-Z_0-9\-\.]+","level":40,"msg":"This is a Standard Log message\\n","name":"test","pid":[0-9]+,"scope":"main","tid":[0-9]+,"time":"[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+Z","topic":"main","v":0}`)
	suite.Assert().Truef(pattern.MatchString(output), "Output is malformed: %s", output)
}

func (suite *LoggerSuite) TestCanUseWithStandardLog() {
	output := CaptureStdout(func() {
		log := logger.Create("test", &logger.StdoutStream{Unbuffered: true}).AsStandardLog()
		log.Print("This is a Standard Log message")
	})
	pattern := regexp.MustCompile(`{"hostname":"[a-zA-Z_0-9\-\.]+","level":30,"msg":"This is a Standard Log message\\n","name":"test","pid":[0-9]+,"scope":"main","tid":[0-9]+,"time":"[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+Z","topic":"main","v":0}`)
	suite.Assert().Truef(pattern.MatchString(output), "Output is malformed: %s", output)
}

func (suite *LoggerSuite) TestCanUseWithStandardLogWithLevel() {
	output := CaptureStdout(func() {
		log := logger.Create("test", &logger.StdoutStream{Unbuffered: true}).AsStandardLog(logger.WARN)
		log.Print("This is a Standard Log message")
	})
	pattern := regexp.MustCompile(`{"hostname":"[a-zA-Z_0-9\-\.]+","level":40,"msg":"This is a Standard Log message\\n","name":"test","pid":[0-9]+,"scope":"main","tid":[0-9]+,"time":"[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+Z","topic":"main","v":0}`)
	suite.Assert().Truef(pattern.MatchString(output), "Output is malformed: %s", output)
}

func (suite *LoggerSuite) TestCanLogAtErrorWithNilError() {
	log, teardown := CreateLogger(suite.T(), "test.log", true)
	defer func() {
		suite.Assert().Nil(recover(), "logger.Errorf did panic")
		teardown()
	}()
	suite.Require().NotNil(log, "cannot create a logger.Logger")

	log.Errorf("Houston, we have a problem", nil)
}

func (suite *LoggerSuite) TestCanLogAtFatalWithNilError() {
	log, teardown := CreateLogger(suite.T(), "test.log", true)
	defer func() {
		suite.Assert().Nil(recover(), "logger.Fatalf did panic")
		teardown()
	}()
	suite.Require().NotNil(log, "cannot create a logger.Logger")

	log.Fatalf("Houston, we have a problem", nil)
}

func (suite *LoggerSuite) TestCanLogMemory() {
	output := CaptureStdout(func() {
		log := logger.Create("test", &logger.StdoutStream{Unbuffered: true, FilterLevel: logger.TRACE})
		log.Memory()
	})
	suite.Require().NotEmpty(output, "There was no output")
	pattern := regexp.MustCompile(`{"hostname":"[a-zA-Z_0-9\-\.]+","level":10,"msg":"Heap\(Alloc = [0-9]+\.[0-9]{2}[GMK]iB, System = [0-9]+\.[0-9]{2}[GMK]iB\), Stack\(Alloc = [0-9]+\.[0-9]{2}[GMK]iB, System = [0-9]+\.[0-9]{2}[GMK]iB\), NumGC = [0-9]+","name":"test","pid":[0-9]+,"scope":"main","tid":[0-9]+,"time":"[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+Z","topic":"main","v":0}`)
	suite.Assert().Truef(pattern.MatchString(output), "Output is malformed: %s", output)
}

func (suite *LoggerSuite) TestCanLogMemoryWithLevel() {
	output := CaptureStdout(func() {
		log := logger.Create("test", &logger.StdoutStream{Unbuffered: true, FilterLevel: logger.TRACE})
		log.Memoryl(logger.INFO)
	})
	suite.Require().NotEmpty(output, "There was no output")
	pattern := regexp.MustCompile(`{"hostname":"[a-zA-Z_0-9\-\.]+","level":30,"msg":"Heap\(Alloc = [0-9]+\.[0-9]{2}[GMK]iB, System = [0-9]+\.[0-9]{2}[GMK]iB\), Stack\(Alloc = [0-9]+\.[0-9]{2}[GMK]iB, System = [0-9]+\.[0-9]{2}[GMK]iB\), NumGC = [0-9]+","name":"test","pid":[0-9]+,"scope":"main","tid":[0-9]+,"time":"[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+Z","topic":"main","v":0}`)
	suite.Assert().Truef(pattern.MatchString(output), "Output is malformed: %s", output)
}

func (suite *LoggerSuite) TestCanLogMemoryWithMessage() {
	output := CaptureStdout(func() {
		log := logger.Create("test", &logger.StdoutStream{Unbuffered: true, FilterLevel: logger.TRACE})
		log.Memoryf("Text %d:", 2)
	})
	suite.Require().NotEmpty(output, "There was no output")
	pattern := regexp.MustCompile(`{"hostname":"[a-zA-Z_0-9\-\.]+","level":10,"msg":"Text 2: Heap\(Alloc = [0-9]+\.[0-9]{2}[GMK]iB, System = [0-9]+\.[0-9]{2}[GMK]iB\), Stack\(Alloc = [0-9]+\.[0-9]{2}[GMK]iB, System = [0-9]+\.[0-9]{2}[GMK]iB\), NumGC = [0-9]+","name":"test","pid":[0-9]+,"scope":"main","tid":[0-9]+,"time":"[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+Z","topic":"main","v":0}`)
	suite.Assert().Truef(pattern.MatchString(output), "Output is malformed: %s", output)
}

func (suite *LoggerSuite) TestCanLogMemoryWithMessageWithLevelAndMessage() {
	output := CaptureStdout(func() {
		log := logger.Create("test", &logger.StdoutStream{Unbuffered: true, FilterLevel: logger.TRACE})
		log.Memorylf(logger.INFO, "Text %d:", 2)
	})
	pattern := regexp.MustCompile(`{"hostname":"[a-zA-Z_0-9\-\.]+","level":30,"msg":"Text 2: Heap\(Alloc = [0-9]+\.[0-9]{2}[GMK]iB, System = [0-9]+\.[0-9]{2}[GMK]iB\), Stack\(Alloc = [0-9]+\.[0-9]{2}[GMK]iB, System = [0-9]+\.[0-9]{2}[GMK]iB\), NumGC = [0-9]+","name":"test","pid":[0-9]+,"scope":"main","tid":[0-9]+,"time":"[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+Z","topic":"main","v":0}`)
	suite.Assert().Truef(pattern.MatchString(output), "Output is malformed: %s", output)
}

type Customer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (customer Customer) Redact() interface{} {
	return Customer{customer.ID, "REDACTED"}
}

func (suite *LoggerSuite) TestCanRedactSensitiveStruct() {
	customer := Customer{"12345678", "John Doe"}
	output := CaptureStdout(func() {
		log := logger.Create("test", &logger.StdoutStream{Unbuffered: true})
		log.Record("customer", customer).Infof("message")
	})
	pattern := regexp.MustCompile(`{"customer":{"id":"12345678","name":"REDACTED"},"hostname":"[a-zA-Z_0-9\-\.]+","level":30,"msg":"message","name":"test","pid":[0-9]+,"scope":"main","tid":[0-9]+,"time":"[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+Z","topic":"main","v":0}`)
	suite.Assert().Truef(pattern.MatchString(output), "Output is malformed: %s", output)
}

func (suite *LoggerSuite) TestCanRedactMessage() {
	output := CaptureStdout(func() {
		log := logger.Create(
			"test",
			&logger.StdoutStream{Unbuffered: true},
			core.Must(logger.NewRedactor(`\+[0-9]{11}`)).(*logger.Redactor),
		)
		log.Infof("message with sensitive (+13178723000) data")
	})
	pattern := regexp.MustCompile(`{"hostname":"[a-zA-Z_0-9\-\.]+","level":30,"msg":"message with sensitive \(REDACTED\) data","name":"test","pid":[0-9]+,"scope":"main","tid":[0-9]+,"time":"[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+Z","topic":"main","v":0}`)
	suite.Assert().Truef(pattern.MatchString(output), "Output is malformed: %s", output)
}

func (suite *LoggerSuite) TestCanFilterMore() {
	log := logger.Create("TEST")
	log.FilterMore()
	// We cannot do this unfortunately:
	// suite.Assert().Equal(logger.WARN, log.FilterLevel)
}

func (suite *LoggerSuite) TestCanFilterLess() {
	log := logger.Create("TEST")
	log.FilterLess()
	// We cannot do this unfortunately:
	// suite.Assert().Equal(logger.DEBUG, log.FilterLevel)
}

func (suite *LoggerSuite) TestCanLogAtDifferentLevelsPerTopic() {
	output := CaptureStdout(func() {
		log := logger.Create("test", &logger.StdoutStream{Unbuffered: true})
		log.SetFilterLevelForTopic(logger.DEBUG, "child")
		log = log.Child("child", nil)
		log.Debugf("message")
	})
	suite.Require().NotEmpty(output, "There was no output")
	pattern := regexp.MustCompile(`{"hostname":"[a-zA-Z_0-9\-\.]+","level":20,"msg":"message","name":"test","pid":[0-9]+,"scope":"main","tid":[0-9]+,"time":"[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+Z","topic":"child","v":0}`)
	suite.Assert().Truef(pattern.MatchString(output), "Output is malformed: %s", output)
}

func (suite *LoggerSuite) TestCanLogAtDifferentLevelsPerTopicAndEmptyScope() {
	output := CaptureStdout(func() {
		log := logger.Create("test", &logger.StdoutStream{Unbuffered: true})
		log.SetFilterLevelForTopic(logger.DEBUG, "child")
		log = log.Child("child", "")
		log.Debugf("message")
	})
	suite.Require().NotEmpty(output, "There was no output")
	pattern := regexp.MustCompile(`{"hostname":"[a-zA-Z_0-9\-\.]+","level":20,"msg":"message","name":"test","pid":[0-9]+,"scope":"","tid":[0-9]+,"time":"[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+Z","topic":"child","v":0}`)
	suite.Assert().Truef(pattern.MatchString(output), "Output is malformed: %s", output)
}

func (suite *LoggerSuite) TestCannotLogAtDifferentLevelsWithEmptyTopicAndEmptyScope() {
	output := CaptureStdout(func() {
		log := logger.Create("test", &logger.StdoutStream{Unbuffered: true})
		log.SetFilterLevelForTopic(logger.DEBUG, "child")
		log = log.Child("", "")
		log.Debugf("message")
	})
	suite.Assert().Empty(output, "There was an output")
}

func (suite *LoggerSuite) TestCanLogWithEmptyTopicAndEmptyScope() {
	output := CaptureStdout(func() {
		log := logger.Create("test", &logger.StdoutStream{Unbuffered: true})
		log.SetFilterLevelForTopic(logger.DEBUG, "child")
		log = log.Child("", "")
		log.Infof("message")
	})
	suite.Require().NotEmpty(output, "There was no output")
	pattern := regexp.MustCompile(`{"hostname":"[a-zA-Z_0-9\-\.]+","level":30,"msg":"message","name":"test","pid":[0-9]+,"scope":"","tid":[0-9]+,"time":"[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+Z","topic":"","v":0}`)
	suite.Assert().Truef(pattern.MatchString(output), "Output is malformed: %s", output)
}

func (suite *LoggerSuite) TestCanLogAtDifferentLevelsPerTopicAndScope() {
	output := CaptureStdout(func() {
		log := logger.Create("test", &logger.StdoutStream{Unbuffered: true})
		log.SetFilterLevelForTopicAndScope(logger.DEBUG, "child", "scope")
		log = log.Child("child", "scope")
		log.Debugf("message")
	})
	suite.Require().NotEmpty(output, "There was no output")
	pattern := regexp.MustCompile(`{"hostname":"[a-zA-Z_0-9\-\.]+","level":20,"msg":"message","name":"test","pid":[0-9]+,"scope":"scope","tid":[0-9]+,"time":"[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+Z","topic":"child","v":0}`)
	suite.Assert().Truef(pattern.MatchString(output), "Output is malformed: %s", output)
}

func (suite *LoggerSuite) TestCanLogTimedFunc() {
	output := CaptureStdout(func() {
		log := logger.Create("test", &logger.StdoutStream{Unbuffered: true})
		log.TimeFunc("code", func() {
			log.Infof("message")
			time.Sleep(500 * time.Millisecond)
		})
	})
	suite.Require().NotEmpty(output, "There was no output")
	lines := strings.Split(output, "\n")
	suite.Require().Len(lines, 3, "There should be 3 lines in the log output, found %d", len(lines))

	pattern := regexp.MustCompile(`{"hostname":"[a-zA-Z_0-9\-\.]+","level":30,"msg":"message","name":"test","pid":[0-9]+,"scope":"main","tid":[0-9]+,"time":"[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+Z","topic":"main","v":0}`)
	suite.Assert().Truef(pattern.MatchString(lines[0]), "Line #1: Output is malformed: \n%s", lines[0])

	pattern = regexp.MustCompile(`{"duration":[0-9]+\.[0-9]+,"hostname":"[a-zA-Z_0-9\-\.]+","level":30,"msg":"code\. executed in [0-9]+\.[0-9]+ms","name":"test","pid":[0-9]+,"scope":"main","tid":[0-9]+,"time":"[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+Z","topic":"main","v":0}`)
	suite.Assert().Truef(pattern.MatchString(lines[1]), "Line #2: Output is malformed: \n%s", lines[1])
}

func (suite *LoggerSuite) TestCanLogTimedFuncWithReturnedValue() {
	var result interface{}
	output := CaptureStdout(func() {
		log := logger.Create("test", &logger.StdoutStream{Unbuffered: true})
		code := func() int {
			log.Infof("message")
			time.Sleep(500 * time.Millisecond)
			return 12
		}
		result = log.TimeFuncV("code", func() interface{} {
			return code()
		})
	})
	suite.Assert().NotNil(result, "There was no result")
	suite.Assert().Equal(12, result.(int), "The result is not 12")
	suite.Require().NotEmpty(output, "There was no output")
	lines := strings.Split(output, "\n")
	suite.Require().Len(lines, 3, "There should be 3 lines in the log output, found %d", len(lines))

	pattern := regexp.MustCompile(`{"hostname":"[a-zA-Z_0-9\-\.]+","level":30,"msg":"message","name":"test","pid":[0-9]+,"scope":"main","tid":[0-9]+,"time":"[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+Z","topic":"main","v":0}`)
	suite.Assert().Truef(pattern.MatchString(lines[0]), "Line #1: Output is malformed: \n%s", lines[0])

	pattern = regexp.MustCompile(`{"duration":[0-9]+\.[0-9]+,"hostname":"[a-zA-Z_0-9\-\.]+","level":30,"msg":"code\. executed in [0-9]+\.[0-9]+ms","name":"test","pid":[0-9]+,"scope":"main","tid":[0-9]+,"time":"[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+Z","topic":"main","v":0}`)
	suite.Assert().Truef(pattern.MatchString(lines[1]), "Line #2: Output is malformed: \n%s", lines[1])
}

func (suite *LoggerSuite) TestCanLogTimedFuncWithReturnedError() {
	var err error
	output := CaptureStdout(func() {
		log := logger.Create("test", &logger.StdoutStream{Unbuffered: true})
		err = log.TimeFuncE("code", func() error {
			log.Infof("message")
			time.Sleep(500 * time.Millisecond)
			return errors.New("error")
		})
	})
	suite.Assert().Error(err, "There was no error")
	suite.Require().NotEmpty(output, "There was no output")
	lines := strings.Split(output, "\n")
	suite.Require().Len(lines, 3, "There should be 3 lines in the log output, found %d", len(lines))

	pattern := regexp.MustCompile(`{"hostname":"[a-zA-Z_0-9\-\.]+","level":30,"msg":"message","name":"test","pid":[0-9]+,"scope":"main","tid":[0-9]+,"time":"[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+Z","topic":"main","v":0}`)
	suite.Assert().Truef(pattern.MatchString(lines[0]), "Line #1: Output is malformed: \n%s", lines[0])

	pattern = regexp.MustCompile(`{"duration":[0-9]+\.[0-9]+,"hostname":"[a-zA-Z_0-9\-\.]+","level":30,"msg":"code\. executed in [0-9]+\.[0-9]+ms","name":"test","pid":[0-9]+,"scope":"main","tid":[0-9]+,"time":"[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+Z","topic":"main","v":0}`)
	suite.Assert().Truef(pattern.MatchString(lines[1]), "Line #2: Output is malformed: \n%s", lines[1])
}

func (suite *LoggerSuite) TestCanLogTimedFuncWithReturnedValueAndError() {
	var result interface{}
	var err error
	output := CaptureStdout(func() {
		log := logger.Create("test", &logger.StdoutStream{Unbuffered: true})
		code := func() (int, error) {
			log.Infof("message")
			time.Sleep(500 * time.Millisecond)
			return 12, errors.New("error")
		}
		result, err = log.TimeFuncVE("code", func() (interface{}, error) {
			return code()
		})
	})
	suite.Assert().Error(err, "There was no error")
	suite.Assert().NotNil(result, "There was no result")
	suite.Assert().Equal(12, result.(int), "The result is not 12")
	suite.Require().NotEmpty(output, "There was no output")
	lines := strings.Split(output, "\n")
	suite.Require().Len(lines, 3, "There should be 3 lines in the log output, found %d", len(lines))

	pattern := regexp.MustCompile(`{"hostname":"[a-zA-Z_0-9\-\.]+","level":30,"msg":"message","name":"test","pid":[0-9]+,"scope":"main","tid":[0-9]+,"time":"[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+Z","topic":"main","v":0}`)
	suite.Assert().Truef(pattern.MatchString(lines[0]), "Line #1: Output is malformed: \n%s", lines[0])

	pattern = regexp.MustCompile(`{"duration":[0-9]+\.[0-9]+,"hostname":"[a-zA-Z_0-9\-\.]+","level":30,"msg":"code\. executed in [0-9]+\.[0-9]+ms","name":"test","pid":[0-9]+,"scope":"main","tid":[0-9]+,"time":"[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+Z","topic":"main","v":0}`)
	suite.Assert().Truef(pattern.MatchString(lines[1]), "Line #2: Output is malformed: \n%s", lines[1])
}
