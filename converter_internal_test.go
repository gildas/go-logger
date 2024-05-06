package logger

import (
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"cloud.google.com/go/logging"
	"github.com/stretchr/testify/suite"
)

type ConverterSuite struct {
	suite.Suite
	Name string
}

func TestConverterSuite(t *testing.T) {
	suite.Run(t, new(ConverterSuite))
}

func (suite *ConverterSuite) SetupSuite() {
	suite.Name = strings.TrimSuffix(reflect.TypeOf(suite).Elem().Name(), "Suite")
}

func (suite *ConverterSuite) TestCanGetConverterFromEnvironment() {
	if current, ok := os.LookupEnv("LOG_CONVERTER"); ok {
		defer os.Setenv("LOG_CONVERTER", current)
	} else {
		defer os.Unsetenv("LOG_CONVERTER")
	}
	os.Setenv("LOG_CONVERTER", "bunyan")
	converter := GetConverterFromEnvironment()
	suite.Assert().IsType(&BunyanConverter{}, converter)
	os.Setenv("LOG_CONVERTER", "aws")
	converter = GetConverterFromEnvironment()
	suite.Assert().IsType(&CloudWatchConverter{}, converter)
	os.Setenv("LOG_CONVERTER", "cloudwatch")
	converter = GetConverterFromEnvironment()
	suite.Assert().IsType(&CloudWatchConverter{}, converter)
	os.Setenv("LOG_CONVERTER", "stackdriver")
	converter = GetConverterFromEnvironment()
	suite.Assert().IsType(&StackDriverConverter{}, converter)
	os.Setenv("LOG_CONVERTER", "google")
	converter = GetConverterFromEnvironment()
	suite.Assert().IsType(&StackDriverConverter{}, converter)
	os.Setenv("LOG_CONVERTER", "gcp")
	converter = GetConverterFromEnvironment()
	suite.Assert().IsType(&StackDriverConverter{}, converter)
	os.Setenv("LOG_CONVERTER", "bello")
	converter = GetConverterFromEnvironment()
	suite.Assert().IsType(&BunyanConverter{}, converter)
}

func (suite *ConverterSuite) TestCanConvertWithBunyanConverter() {
	converter := &BunyanConverter{}
	record := NewRecord().Set("time", time.Now().UTC()).Set("bello", "banana")
	converted := converter.Convert(record)
	suite.Assert().Exactly(record, converted)
	value := converted.Get("time")
	suite.Require().IsType("string", value)
	stime := value.(string)
	suite.Assert().Regexp(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z`, stime)
}

func (suite *ConverterSuite) TestCanConvertWithPinoConverter() {
	converter := &PinoConverter{}
	record := NewRecord().Set("time", time.Now().UTC()).Set("level", INFO).Set("bello", "banana")
	converted := converter.Convert(record)
	suite.Assert().IsType(int64(0), converted.Get("time"))
	suite.Assert().Equal(30, converted.Get("level"))
	suite.Assert().Equal(1, converted.Get("v"))
	// {"level":30,"time":1573664685469,"bello":"banana","v":1}
}

func (suite *ConverterSuite) TestCanConvertWithStackDriverConverter() {
	now := time.Now().UTC()
	converter := &StackDriverConverter{}
	record := converter.Convert(NewRecord().Set("level", INFO).Set("time", now).Set("msg", "Hello World!").Set("v", 0))
	suite.Assert().Contains(record.Data, "message")
	suite.Assert().NotContains(record.Data, "msg")
	suite.Assert().Contains(record.Data, "level")
	suite.Assert().Contains(record.Data, "severity")
	suite.Assert().Contains(record.Data, "time")
	value := record.Get("time")
	suite.Require().IsType("string", value)
	stime := value.(string)
	suite.Assert().Regexp(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z`, stime)
	suite.Assert().Contains(record.Data, "v")
	suite.Assert().NotContains(record.Data, "name")

	record = converter.Convert(NewRecord().Set("level", NEVER).Set("msg", "Hello World!"))
	suite.Assert().Equal(logging.Info, record.Get("severity"))

	record = converter.Convert(NewRecord().Set("level", TRACE).Set("msg", "Hello World!"))
	suite.Assert().Equal(logging.Debug, record.Get("severity"))

	record = converter.Convert(NewRecord().Set("level", DEBUG).Set("msg", "Hello World!"))
	suite.Assert().Equal(logging.Debug, record.Get("severity"))

	record = converter.Convert(NewRecord().Set("level", WARN).Set("msg", "Hello World!"))
	suite.Assert().Equal(logging.Warning, record.Get("severity"))

	record = converter.Convert(NewRecord().Set("level", ERROR).Set("msg", "Hello World!"))
	suite.Assert().Equal(logging.Error, record.Get("severity"))

	record = converter.Convert(NewRecord().Set("level", FATAL).Set("msg", "Hello World!"))
	suite.Assert().Equal(logging.Critical, record.Get("severity"))

	record = converter.Convert(NewRecord().Set("level", ALWAYS).Set("msg", "Hello World!"))
	suite.Assert().Equal(logging.Emergency, record.Get("severity"))

	record = converter.Convert(NewRecord().Set("msg", "Hello World!"))
	suite.Assert().Equal(logging.Info, record.Get("severity"))
}

func (suite *ConverterSuite) TestCanConvertWithCloudWatchConverter() {
	now := time.Now().UTC()
	converter := &CloudWatchConverter{}
	record := converter.Convert(NewRecord().Set("level", INFO).Set("time", now).Set("msg", "Hello World!").Set("v", 0))
	suite.Assert().Contains(record.Data, "level")
	suite.Assert().Contains(record.Data, "msg")
	suite.Assert().Contains(record.Data, "time")
	suite.Assert().IsType("string", record.Get("time"))
	suite.Assert().Contains(record.Data, "v")
	suite.Assert().NotContains(record.Data, "name")

	record = converter.Convert(NewRecord().Set("level", NEVER).Set("msg", "Hello World!"))
	suite.Assert().Equal(NEVER.String(), record.Get("severity"))
	suite.Assert().Equal(NEVER, record.Get("level"))

	record = converter.Convert(NewRecord().Set("level", TRACE).Set("msg", "Hello World!"))
	suite.Assert().Equal(TRACE.String(), record.Get("severity"))
	suite.Assert().Equal(TRACE, record.Get("level"))

	record = converter.Convert(NewRecord().Set("level", DEBUG).Set("msg", "Hello World!"))
	suite.Assert().Equal(DEBUG.String(), record.Get("severity"))
	suite.Assert().Equal(DEBUG, record.Get("level"))

	record = converter.Convert(NewRecord().Set("level", INFO).Set("msg", "Hello World!"))
	suite.Assert().Equal(INFO.String(), record.Get("severity"))
	suite.Assert().Equal(INFO, record.Get("level"))

	record = converter.Convert(NewRecord().Set("level", WARN).Set("msg", "Hello World!"))
	suite.Assert().Equal(WARN.String(), record.Get("severity"))
	suite.Assert().Equal(WARN, record.Get("level"))

	record = converter.Convert(NewRecord().Set("level", ERROR).Set("msg", "Hello World!"))
	suite.Assert().Equal(ERROR.String(), record.Get("severity"))
	suite.Assert().Equal(ERROR, record.Get("level"))

	record = converter.Convert(NewRecord().Set("level", FATAL).Set("msg", "Hello World!"))
	suite.Assert().Equal(FATAL.String(), record.Get("severity"))
	suite.Assert().Equal(FATAL, record.Get("level"))

	record = converter.Convert(NewRecord().Set("level", ALWAYS).Set("msg", "Hello World!"))
	suite.Assert().Equal(ALWAYS.String(), record.Get("severity"))
	suite.Assert().Equal(ALWAYS, record.Get("level"))
}
