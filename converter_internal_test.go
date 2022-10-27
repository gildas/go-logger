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
	suite.Assert().IsType("string", converted["time"])
}

func (suite *ConverterSuite) TestCanConvertWithPinoConverter() {
	converter := &PinoConverter{}
	record := NewRecord().Set("time", time.Now().UTC()).Set("level", INFO).Set("bello", "banana")
	converted := converter.Convert(record)
	suite.Assert().IsType(int64(0), converted["time"])
	suite.Assert().Equal(30, converted["level"])
	suite.Assert().Equal(1, converted["v"])
	// {"level":30,"time":1573664685469,"bello":"banana","v":1}
}

func (suite *ConverterSuite) TestCanConvertWithStackDriverConverter() {
	now := time.Now().UTC()
	converter := &StackDriverConverter{}
	record := converter.Convert(NewRecord().Set("level", INFO).Set("time", now).Set("msg", "Hello World!").Set("v", 0))
	suite.Assert().Contains(record, "message")
	suite.Assert().Contains(record, "msg")
	suite.Assert().Contains(record, "level")
	suite.Assert().Contains(record, "severity")
	suite.Assert().Contains(record, "time")
	suite.Assert().IsType("string", record["time"])
	suite.Assert().Contains(record, "v")
	suite.Assert().NotContains(record, "name")

	record = converter.Convert(NewRecord().Set("level", NEVER).Set("msg", "Hello World!"))
	suite.Assert().Equal(logging.Info, record["severity"])

	record = converter.Convert(NewRecord().Set("level", TRACE).Set("msg", "Hello World!"))
	suite.Assert().Equal(logging.Debug, record["severity"])

	record = converter.Convert(NewRecord().Set("level", DEBUG).Set("msg", "Hello World!"))
	suite.Assert().Equal(logging.Debug, record["severity"])

	record = converter.Convert(NewRecord().Set("level", WARN).Set("msg", "Hello World!"))
	suite.Assert().Equal(logging.Warning, record["severity"])

	record = converter.Convert(NewRecord().Set("level", ERROR).Set("msg", "Hello World!"))
	suite.Assert().Equal(logging.Error, record["severity"])

	record = converter.Convert(NewRecord().Set("level", FATAL).Set("msg", "Hello World!"))
	suite.Assert().Equal(logging.Critical, record["severity"])

	record = converter.Convert(NewRecord().Set("level", ALWAYS).Set("msg", "Hello World!"))
	suite.Assert().Equal(logging.Emergency, record["severity"])

	record = converter.Convert(NewRecord().Set("msg", "Hello World!"))
	suite.Assert().Equal(logging.Info, record["severity"])
}

func (suite *ConverterSuite) TestCanConvertWithCloudWatchConverter() {
	now := time.Now().UTC()
	converter := &CloudWatchConverter{}
	record := converter.Convert(NewRecord().Set("level", INFO).Set("time", now).Set("msg", "Hello World!").Set("v", 0))
	suite.Assert().Contains(record, "level")
	suite.Assert().Contains(record, "msg")
	suite.Assert().Contains(record, "time")
	suite.Assert().IsType("string", record["time"])
	suite.Assert().Contains(record, "v")
	suite.Assert().NotContains(record, "name")

	record = converter.Convert(NewRecord().Set("level", NEVER).Set("msg", "Hello World!"))
	suite.Assert().Equal(NEVER.String(), record["severity"])
	suite.Assert().Equal(NEVER, record["level"])

	record = converter.Convert(NewRecord().Set("level", TRACE).Set("msg", "Hello World!"))
	suite.Assert().Equal(TRACE.String(), record["severity"])
	suite.Assert().Equal(TRACE, record["level"])

	record = converter.Convert(NewRecord().Set("level", DEBUG).Set("msg", "Hello World!"))
	suite.Assert().Equal(DEBUG.String(), record["severity"])
	suite.Assert().Equal(DEBUG, record["level"])

	record = converter.Convert(NewRecord().Set("level", INFO).Set("msg", "Hello World!"))
	suite.Assert().Equal(INFO.String(), record["severity"])
	suite.Assert().Equal(INFO, record["level"])

	record = converter.Convert(NewRecord().Set("level", WARN).Set("msg", "Hello World!"))
	suite.Assert().Equal(WARN.String(), record["severity"])
	suite.Assert().Equal(WARN, record["level"])

	record = converter.Convert(NewRecord().Set("level", ERROR).Set("msg", "Hello World!"))
	suite.Assert().Equal(ERROR.String(), record["severity"])
	suite.Assert().Equal(ERROR, record["level"])

	record = converter.Convert(NewRecord().Set("level", FATAL).Set("msg", "Hello World!"))
	suite.Assert().Equal(FATAL.String(), record["severity"])
	suite.Assert().Equal(FATAL, record["level"])

	record = converter.Convert(NewRecord().Set("level", ALWAYS).Set("msg", "Hello World!"))
	suite.Assert().Equal(ALWAYS.String(), record["severity"])
	suite.Assert().Equal(ALWAYS, record["level"])
}
