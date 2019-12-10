package logger

import (
	"reflect"
	"strings"
	"testing"

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

func (suite *ConverterSuite) TestCanConvertWithBunyanConverter() {
	converter := &BunyanConverter{}
	record := NewRecord().Set("bello", "banana")
	converted := converter.Convert(record)
	suite.Assert().Exactly(record, converted)
}

func (suite *ConverterSuite) TestCanConvertWithStackDriverConverter() {
	converter := &StackDriverConverter{}
	record := converter.Convert(NewRecord().Set("level", INFO).Set("msg", "Hello World!"))
	suite.Assert().Contains(record, "message")
	suite.Assert().Contains(record, "severity")
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