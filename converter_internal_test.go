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

func (suite *ConverterSuite) TestCanGetConverterFromEnvironment() {
	os.Setenv("LOG_CONVERTER", "bunyan")
	converter := GetConverterFromEnvironment()
	suite.Assert().IsType(&BunyanConverter{}, converter)
	os.Setenv("LOG_CONVERTER", "stackdriver")
	converter = GetConverterFromEnvironment()
	suite.Assert().IsType(&StackDriverConverter{}, converter)
	os.Setenv("LOG_CONVERTER", "bello")
	converter = GetConverterFromEnvironment()
	suite.Assert().IsType(&BunyanConverter{}, converter)
	os.Unsetenv("LOG_CONVERTER")
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
	converter := &StackDriverConverter{}
	record := converter.Convert(NewRecord().Set("level", INFO).Set("time", time.Now().UTC()).Set("msg", "Hello World!"))
	suite.Assert().Contains(record, "message")
	suite.Assert().Contains(record, "severity")
	suite.Assert().Contains(record, "time")
	suite.Assert().IsType("string", record["time"])
	suite.Assert().NotContains(record, "level")
	suite.Assert().NotContains(record, "msg")
	suite.Assert().NotContains(record, "name")
	suite.Assert().NotContains(record, "v")

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

func (suite *ConverterSuite) SetupSuite() {
	suite.Name = strings.TrimSuffix(reflect.TypeOf(*suite).Name(), "Suite")
}
