package logger_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
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
	expected := `{"bello": "banana", "だれ": "私"}`
	me := "私"
	record := logger.NewRecord().Set("bello", "banana").Set("だれ", &me)

	suite.Require().Len(record.Data, 2)

	payload, err := json.Marshal(record)
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(expected, string(payload))

	payload, err = json.Marshal(logger.NewRecord())
	suite.Require().NoError(err, "Error while marshaling empty record")
	suite.Assert().JSONEq("null", string(payload))
}

func (suite *RecordSuite) TestCanMarshalEmpty() {
	record := logger.NewRecord()
	payload, err := json.Marshal(record)
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(`null`, string(payload))
}

func (suite *RecordSuite) TestCanMarshalNilValue() {
	record := logger.NewRecord()
	record.Data["empty"] = nil
	payload, err := json.Marshal(record)
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(`{}`, string(payload))

	record.Data["?empty"] = nil
	payload, err = json.Marshal(record)
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(`{"empty": null}`, string(payload))
}

func (suite *RecordSuite) TestCanMarshalEmptyString() {
	record := logger.NewRecord()
	record.Data["empty"] = ""
	payload, err := json.Marshal(record)
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(`{}`, string(payload))

	record.Data["?empty"] = ""
	payload, err = json.Marshal(record)
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(`{"empty": ""}`, string(payload))
}

func (suite *RecordSuite) TestCanMarshalEmptyUUID() {
	record := logger.NewRecord()
	record.Data["empty"] = uuid.Nil
	payload, err := json.Marshal(record)
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(`{}`, string(payload))

	record.Data["?empty"] = uuid.Nil
	payload, err = json.Marshal(record)
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(`{"empty": "00000000-0000-0000-0000-000000000000"}`, string(payload))
}

func (suite *RecordSuite) TestCanMarshalIsNil() {
	record := logger.NewRecord()
	record.Data["empty"] = User{}
	payload, err := json.Marshal(record)
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(`{}`, string(payload))

	record.Data["?empty"] = User{}
	payload, err = json.Marshal(record)
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(`{"empty": {"id": "", "name": ""}}`, string(payload))
}

func (suite *RecordSuite) TestCanMarshalStringValueWithSpecialCharacters() {
	expected := `{"key": "\b\f\t\n\r\\\"Hello\""}`

	payload, err := json.Marshal(logger.NewRecord().Set("key", "\b\f\t\n\r\\\"Hello\""))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(expected, string(payload))
}

func (suite *RecordSuite) TestCanMarshalBasicValues() {
	expected := `{"key": 123}`

	payload, err := json.Marshal(logger.NewRecord().Set("key", true))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(`{"key": true}`, string(payload))

	payload, err = json.Marshal(logger.NewRecord().Set("key", 123))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(expected, string(payload))

	payload, err = json.Marshal(logger.NewRecord().Set("key", int8(123)))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(expected, string(payload))

	payload, err = json.Marshal(logger.NewRecord().Set("key", int16(123)))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(expected, string(payload))

	payload, err = json.Marshal(logger.NewRecord().Set("key", int32(123)))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(expected, string(payload))

	payload, err = json.Marshal(logger.NewRecord().Set("key", int64(123)))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(expected, string(payload))

	payload, err = json.Marshal(logger.NewRecord().Set("key", uint(123)))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(expected, string(payload))

	payload, err = json.Marshal(logger.NewRecord().Set("key", uint8(123)))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(expected, string(payload))

	payload, err = json.Marshal(logger.NewRecord().Set("key", uint16(123)))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(expected, string(payload))

	payload, err = json.Marshal(logger.NewRecord().Set("key", uint32(123)))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(expected, string(payload))

	payload, err = json.Marshal(logger.NewRecord().Set("key", uint64(123)))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(expected, string(payload))

	payload, err = json.Marshal(logger.NewRecord().Set("key", float32(123)))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(expected, string(payload))

	payload, err = json.Marshal(logger.NewRecord().Set("key", float64(123)))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(expected, string(payload))

	payload, err = json.Marshal(logger.NewRecord().Set("key", complex64(123+4i)))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(`{"key": "(123+4i)"}`, string(payload))

	payload, err = json.Marshal(logger.NewRecord().Set("key", complex128(123+4i)))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(`{"key": "(123+4i)"}`, string(payload))
}

func (suite *RecordSuite) TestCanMarshalPointerToBasicValues() {
	expected := `{"key": 123}`

	bvalue := true
	payload, err := json.Marshal(logger.NewRecord().Set("key", &bvalue))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(`{"key": true}`, string(payload))

	ivalue := 123
	payload, err = json.Marshal(logger.NewRecord().Set("key", &ivalue))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(expected, string(payload))

	i8value := int8(123)
	payload, err = json.Marshal(logger.NewRecord().Set("key", &i8value))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(expected, string(payload))

	i16value := int16(123)
	payload, err = json.Marshal(logger.NewRecord().Set("key", &i16value))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(expected, string(payload))

	i32value := int32(123)
	payload, err = json.Marshal(logger.NewRecord().Set("key", &i32value))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(expected, string(payload))

	i64value := int64(123)
	payload, err = json.Marshal(logger.NewRecord().Set("key", &i64value))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(expected, string(payload))

	uivalue := uint(123)
	payload, err = json.Marshal(logger.NewRecord().Set("key", &uivalue))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(expected, string(payload))

	ui8value := uint8(123)
	payload, err = json.Marshal(logger.NewRecord().Set("key", &ui8value))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(expected, string(payload))

	ui16value := uint16(123)
	payload, err = json.Marshal(logger.NewRecord().Set("key", &ui16value))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(expected, string(payload))

	ui32value := uint32(123)
	payload, err = json.Marshal(logger.NewRecord().Set("key", &ui32value))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(expected, string(payload))

	ui64value := uint64(123)
	payload, err = json.Marshal(logger.NewRecord().Set("key", &ui64value))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(expected, string(payload))

	f32value := float32(123)
	payload, err = json.Marshal(logger.NewRecord().Set("key", &f32value))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(expected, string(payload))

	f64value := float64(123)
	payload, err = json.Marshal(logger.NewRecord().Set("key", &f64value))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(expected, string(payload))

	c64value := complex64(123 + 4i)
	payload, err = json.Marshal(logger.NewRecord().Set("key", &c64value))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(`{"key": "(123+4i)"}`, string(payload))

	c128value := complex128(123 + 4i)
	payload, err = json.Marshal(logger.NewRecord().Set("key", &c128value))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(`{"key": "(123+4i)"}`, string(payload))
}

func (suite *RecordSuite) TestCanMarshalError() {
	expected := `{"key": {"code": 500, "id": "error.runtime", "text": "banana", "type": "error"}}`
	payload, err := json.Marshal(logger.NewRecord().Set("key", errors.New("banana")))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(expected, string(payload))
}

func (suite *RecordSuite) TestCanMarshalNonMarshableError() {
	expected := `{"key": "banana"}`
	payload, err := json.Marshal(logger.NewRecord().Set("key", NonMarshableError{Message: "banana", Channel: make(chan int)}))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(expected, string(payload))
}

func (suite *RecordSuite) TestCanMarshalNonMarshableObject() {
	expected := `{"key": "banana"}`
	payload, err := json.Marshal(logger.NewRecord().Set("key", NonMarshableObject{Message: "banana", Channel: make(chan int)}))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(expected, string(payload))
}

func (suite *RecordSuite) TestCanMarshalBogusValue() {
	expected := `{"key": "logger_test.BogusValue{}"}`
	payload, err := json.Marshal(logger.NewRecord().Set("key", BogusValue{}))
	suite.Require().NoError(err, "Error while marshaling record")
	suite.Assert().JSONEq(expected, string(payload))
}

func (suite *RecordSuite) TestCanUnmarshal() {
	source := `{"bello": "banana", "だれ": "私"}`
	record := logger.NewRecord()
	err := json.Unmarshal([]byte(source), &record)
	suite.Require().NoError(err, "Error while unmarshaling record")
	suite.Require().Len(record.Data, 2)
	suite.Assert().Contains(record.Data, "bello")
	suite.Assert().Contains(record.Data, "だれ")
	value, found := record.Find("bello")
	suite.Require().True(found, `Record["bello"] should be found`)
	str, ok := value.(string)
	suite.Require().True(ok, `Record["bello"] should be a string`)
	suite.Assert().Equal("banana", str, `Record["bello"] should be "banana"`)
}

func (suite *RecordSuite) TestFailsUnmarshalInvalidJSON() {
	source := `[]`
	record := logger.NewRecord()
	err := json.Unmarshal([]byte(source), &record)
	suite.Require().Error(err)
	suite.Assert().True(errors.Is(err, errors.JSONUnmarshalError), "error should be a JSON Unmarshal error")
}

func (suite *RecordSuite) TestCanSet() {
	record := logger.NewRecord().Set("key", "value")
	suite.Require().NotNil(record, "Failed to create a Record")
	suite.Require().Len(record.Data, 1)
	suite.Assert().Contains(record.Data, "key")
	value, ok := record.Find("key")
	suite.Require().True(ok, `Record["key"] should be found`)
	str, ok := value.(string)
	suite.Require().True(ok, `Record["key"] should be a string`)
	suite.Assert().Equal("value", str, `Record["key"] should be "value"`)
	record.Set("nilvalue", nil)
	suite.Assert().NotContains(record.Data, "nilvalue")
}

func (suite *RecordSuite) TestCannotOverwrite() {
	record := logger.NewRecord().Set("key", "value")
	suite.Require().NotNil(record, "Failed to create a Record")
	suite.Require().Len(record.Data, 1)
	suite.Assert().Contains(record.Data, "key")
	value, found := record.Find("key")
	suite.Require().True(found, `Record["key"] should be a string`)
	str, ok := value.(string)
	suite.Require().True(ok, `Record["key"] should be a string`)
	suite.Assert().Equal("value", str, `Record["key"] should be "value"`)

	record.Set("key", "value2")
	value, found = record.Find("key")
	suite.Require().True(found, `Record["key"] should be found`)
	str, ok = value.(string)
	suite.Require().True(ok, `Record["key"] should be a string`)
	suite.Assert().NotEqual("value2", str, `Record["key"] should not be "value2"`)
}

func (suite *RecordSuite) TestCanMerge() {
	record := logger.NewRecord().Set("bello", "banana").Set("だれ", "私")
	source := logger.NewRecord().Set("bello", "beebop").Set("Para Me", "okido")
	record.Merge(source)
	suite.Require().Len(record.Data, 3)
	suite.Assert().Contains(record.Data, "bello")
	suite.Assert().Contains(record.Data, "だれ")
	suite.Assert().Contains(record.Data, "Para Me")

	record.Merge(nil)
	suite.Require().Len(record.Data, 3)
}

func (suite *RecordSuite) TestCannotFindValueFromNilRecord() {
	var record *logger.Record
	value, found := record.Find("key")
	suite.Require().False(found, `Record["key"] should not be found`)
	suite.Assert().Nil(value, `Record["key"] should be nil`)
}

func (suite *RecordSuite) TestCanAddKeysToRedact() {
	record := logger.NewRecord()
	record.AddKeysToRedact("key1", "key2", "key3")
	suite.Assert().Len(record.KeysToRedact, 3)
	suite.Assert().Contains(record.KeysToRedact, "key1")
	suite.Assert().Contains(record.KeysToRedact, "key2")
	suite.Assert().Contains(record.KeysToRedact, "key3")
	suite.Assert().NotContains(record.KeysToRedact, "key4")
}
