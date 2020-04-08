package logger_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/gildas/go-errors"
	"github.com/gildas/go-logger"
)

type RecordSuite struct {
	suite.Suite
}

func TestRecordSuite(t *testing.T) {
	suite.Run(t, new(RecordSuite))
}
func (suite *RecordSuite) TestCanInstantiate() {
	record := logger.NewRecord()
	suite.Require().NotNil(record, "Failed to create a Record")
}

func (suite *RecordSuite) TestCanMarshal() {
	expected := `{"bello": "banana", "だれ": "Me"}`
	record := logger.NewRecord().Set("bello", "banana").Set("だれ", "Me")

	suite.Require().Len(record, 2)

	payload, err := json.Marshal(record)
	suite.Require().Nil(err, "Error while marshaling record")
	suite.Assert().JSONEq(expected, string(payload))

	payload, err = json.Marshal(logger.NewRecord())
	suite.Require().Nil(err, "Error while marshaling empty record")
	suite.Assert().JSONEq("null", string(payload))
}

func (suite *RecordSuite) TestCanUnmarshal() {
	source := `{"bello": "banana", "だれ": "Me"}`
	record := logger.NewRecord()
	err := json.Unmarshal([]byte(source), &record)
	suite.Require().Nil(err, "Error while unmarshaling record")
	suite.Require().Len(record, 2)
	suite.Assert().Contains(record, "bello")
	suite.Assert().Contains(record, "だれ")
	value, ok := record["bello"].(string)
	suite.Require().True(ok, `Record["bello"] should be a string`)
	suite.Assert().Equal("banana", value, `Record["key"] should be "banana"`)
}

func (suite *RecordSuite) TestFailsUnmarshalInvalidJSON() {
	source := `[]`
	record := logger.NewRecord()
	err := json.Unmarshal([]byte(source), &record)
	suite.Require().NotNil(err)
	suite.Assert().True(errors.Is(err, errors.JSONUnmarshalError), "error should be a JSON Unmarshal error")
}

func (suite *RecordSuite) TestCanSet() {
	record := logger.NewRecord().Set("key", "value")
	suite.Require().NotNil(record, "Failed to create a Record")
	suite.Require().Len(record, 1)
	suite.Assert().Contains(record, "key")
	value, ok := record["key"].(string)
	suite.Require().True(ok, `Record["key"] should be a string`)
	suite.Assert().Equal("value", value, `Record["key"] should be "value"`)
	record.Set("nilvalue", nil)
	suite.Assert().NotContains(record, "nilvalue")
}

func (suite *RecordSuite) TestCannotOverwrite() {
	record := logger.NewRecord().Set("key", "value")
	suite.Require().NotNil(record, "Failed to create a Record")
	suite.Require().Len(record, 1)
	suite.Assert().Contains(record, "key")
	value, ok := record["key"].(string)
	suite.Require().True(ok, `Record["key"] should be a string`)
	suite.Assert().Equal("value", value, `Record["key"] should be "value"`)

	record.Set("key", "value2")
	value, ok = record["key"].(string)
	suite.Require().True(ok, `Record["key"] should be a string`)
	suite.Assert().NotEqual("value2", value, `Record["key"] should not be "value2"`)
}

func (suite *RecordSuite) TestCanMerge() {
	record := logger.NewRecord().Set("bello", "banana").Set("だれ", "Me")
	source := logger.NewRecord().Set("bello", "beebop").Set("Para Me", "okido")
	record.Merge(source)
	suite.Require().Len(record, 3)
	suite.Assert().Contains(record, "bello")
	suite.Assert().Contains(record, "だれ")
	suite.Assert().Contains(record, "Para Me")

	record.Merge(nil)
	suite.Require().Len(record, 3)
}
