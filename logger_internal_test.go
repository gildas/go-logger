package logger

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gildas/go-core"
	"github.com/stretchr/testify/suite"
)

type InternalLoggerSuite struct {
	suite.Suite
	Name string
}

func TestInternalLoggerSuite(t *testing.T) {
	suite.Run(t, new(InternalLoggerSuite))
}

func (suite *InternalLoggerSuite) TestCanCreate() {
	log := Create("test")
	suite.Require().NotNil(log, "Failed to create a Logger with default options")
	suite.Assert().IsType(&StdoutStream{}, log.stream)
	suite.Assert().Equal(false, log.stream.(*StdoutStream).Unbuffered, "stream should be buffered")
	suite.Assert().Equal(INFO, log.stream.(*StdoutStream).FilterLevel, "FilterLevel should be INFO")
}

func (suite *InternalLoggerSuite) TestCanCreateWithStream() {
	log := Create("test", &NilStream{})
	suite.Require().NotNil(log, "Failed to create a Logger with a nil stream")
	suite.Assert().IsType(&NilStream{}, log.stream)
}

func (suite *InternalLoggerSuite) TestCanCreateWithFileStream() {
	log := Create("test", &FileStream{Path: "/var/log/test.log"})
	suite.Assert().IsType(&FileStream{}, log.stream)
	suite.Assert().Equal("/var/log/test.log", log.stream.(*FileStream).Path)
	suite.Assert().Equal(false, log.stream.(*FileStream).Unbuffered, "FileStream should be buffered")
}

func (suite *InternalLoggerSuite) TestCanCreateWithMultipleStreams() {
	log := CreateWithStream("test", &StdoutStream{}, &StackDriverStream{})
	suite.Require().NotNil(log, "Failed to create a Logger with 2 streams")
	suite.Assert().IsType(&MultiStream{}, log.stream)
	suite.Require().Len(log.stream.(*MultiStream).streams, 2)
	suite.Assert().IsType(&StdoutStream{}, log.stream.(*MultiStream).streams[0])
	suite.Assert().IsType(&StackDriverStream{}, log.stream.(*MultiStream).streams[1])
}

func (suite *InternalLoggerSuite) TestCanCreateWithFilterLevel() {
	log := Create("test", TRACE)
	suite.Require().NotNil(log, "Failed to create a Logger")
	suite.Assert().IsType(&StdoutStream{}, log.stream)
	suite.Assert().Equal(TRACE, log.stream.(*StdoutStream).FilterLevel, "FilterLevel should be TRACE")
}

func (suite *InternalLoggerSuite) TestCanCreateWithEnvironmentDEBUG() {
	os.Setenv("DEBUG", "1")
	defer os.Unsetenv("DEBUG")
	log := Create("test")
	suite.Require().NotNil(log, "Failed to create a Logger with stdout stream")
	suite.Assert().IsType(&StdoutStream{}, log.stream)
	suite.Assert().Equal(true, log.stream.(*StdoutStream).Unbuffered, "In DEBUG mode, stdout should be unbuffered")
	suite.Assert().Equal(DEBUG, log.stream.(*StdoutStream).FilterLevel, "FilterLevel should be DEBUG")
}

func (suite *InternalLoggerSuite) TestCanCreateWithEnvironmentFLUSHFREQUENCY() {
	os.Setenv("LOG_FLUSHFREQUENCY", "10ms")
	defer os.Unsetenv("LOG_FLUSHFREQUENCY")
	log := Create("test")
	suite.Require().NotNil(log, "Failed to create a Logger with stdout stream")
	suite.Assert().IsType(&StdoutStream{}, log.stream)
	suite.Assert().Equal(false, log.stream.(*StdoutStream).Unbuffered, "stream should be buffered")
	suite.Assert().Equal(INFO, log.stream.(*StdoutStream).FilterLevel, "FilterLevel should be INFO")
	_ = captureStdout(func() {
		log.Infof("writing something")
	})
	suite.Assert().Equal(10*time.Millisecond, log.stream.(*StdoutStream).flushFrequency, "this stream should flush every 10 milliseconds")
}

func (suite *InternalLoggerSuite) TestCanCreateWithEnvironmentDESTINATION() {
	os.Setenv("LOG_DESTINATION", "/var/log/test.log")
	defer os.Unsetenv("LOG_DESTINATION")
	log := Create("test")
	suite.Require().NotNil(log, "Failed to create a Logger with file stream")
	suite.Assert().IsType(&FileStream{}, log.stream)
	suite.Assert().Equal("/var/log/test.log", log.stream.(*FileStream).Path)
	suite.Assert().Equal(false, log.stream.(*FileStream).Unbuffered, "FileStream should be buffered")
}

func (suite *InternalLoggerSuite) TestCanCreateWithDestination() {
	var log *Logger

	log = CreateWithDestination("test")
	suite.Require().NotNil(log, "cannot create a Logger with no destination")
	suite.Assert().IsType(&StdoutStream{}, log.stream)

	log = CreateWithDestination("test", "/var/log/test.log")
	suite.Require().NotNil(log, "cannot create a Logger with a destination")
	suite.Assert().IsType(&FileStream{}, log.stream)

	log = CreateWithDestination("test", "/var/log/test.log", "stackdriver")
	suite.Require().NotNil(log, "cannot create a Logger with 2 destinations")
	suite.Require().IsType(&MultiStream{}, log.stream)
	suite.Require().Len(log.stream.(*MultiStream).streams, 2)
	suite.Assert().IsType(&FileStream{}, log.stream.(*MultiStream).streams[0])
	suite.Assert().IsType(&StackDriverStream{}, log.stream.(*MultiStream).streams[1])
}

func (suite *InternalLoggerSuite) TestCanCreateWithRedactors() {
	log := Create(
		"test",
		*(core.Must(NewRedactor("[0-9]{8}")).(*Redactor)),
		core.Must(NewRedactor("[a-z]{8}")).(*Redactor),
	)
	suite.Require().NotNil(log, "cannot create a Logger with redactors")
	suite.Require().Len(log.redactors, 2, "The Logger should have 2 redactors")
}

func (suite *InternalLoggerSuite) TestCanCreateChildLoggerWithRedactors() {
	log := Create(
		"test",
		core.Must(NewRedactor("[0-9]{8}")).(*Redactor),
		core.Must(NewRedactor("[a-z]{8}")).(*Redactor),
	)
	suite.Require().NotNil(log, "cannot create a Logger with redactors")
	suite.Require().Len(log.redactors, 2, "The Logger should have 2 redactors")

	child := log.Child(nil, nil, "mytopic", "myvalue", core.Must(NewRedactor("[A-Z]{8}")).(*Redactor))
	suite.Require().NotNil(child, "cannot create a child Logger")
	suite.Assert().Len(child.redactors, 3, "The Child Logger should have 3 redactors")
	suite.Assert().Len(log.redactors, 2, "The Parent Logger should have 2 redactors")

	child2 := log.Child(nil, nil, "mytopic", 3, *(core.Must(NewRedactor("[A-Z]{8}")).(*Redactor)))
	suite.Require().NotNil(child2, "cannot create a child Logger")
	suite.Assert().Len(child.redactors, 3, "The Child Logger should have 3 redactors")
	suite.Assert().Len(log.redactors, 2, "The Parent Logger should have 2 redactors")
}

func (suite *InternalLoggerSuite) TestCanSmartCreateNoParameter() {
	log := Create("test")
	suite.Require().NotNil(log, "cannot create a Logger without parameters")
	suite.Assert().IsType(&StdoutStream{}, log.stream)
}

func (suite *InternalLoggerSuite) TestCanSmartCreateWithNilStream() {
	log := Create("test", &NilStream{})
	suite.Require().NotNil(log, "cannot create a Logger without parameters")
	suite.Assert().IsType(&NilStream{}, log.stream)
}

func (suite *InternalLoggerSuite) TestCanSmartCreateWithNilDestination() {
	log := Create("test", "nil")
	suite.Require().NotNil(log, "cannot create a Logger without parameters")
	suite.Assert().IsType(&NilStream{}, log.stream)
}

func (suite *InternalLoggerSuite) TestCanSmartCreateWithStdout() {
	log := Create("test", &StdoutStream{})
	suite.Require().NotNil(log, "cannot create a Logger with stdout stream")
	suite.Assert().IsType(&StdoutStream{}, log.stream)
}

func (suite *InternalLoggerSuite) TestCanSmartCreateWithStdoutDestination() {
	log := Create("test", "stdout")
	suite.Require().NotNil(log, "cannot create a Logger with stdout destination")
	suite.Assert().IsType(&StdoutStream{}, log.stream)
}

func (suite *InternalLoggerSuite) TestCanSmartCreateWithStderr() {
	log := Create("test", &StderrStream{})
	suite.Require().NotNil(log, "cannot create a Logger with stderr stream")
	suite.Assert().IsType(&StderrStream{}, log.stream)
}

func (suite *InternalLoggerSuite) TestCanSmartCreateWithStderrDestination() {
	log := Create("test", "stderr")
	suite.Require().NotNil(log, "cannot create a Logger with stderr destination")
	suite.Assert().IsType(&StderrStream{}, log.stream)
}

func (suite *InternalLoggerSuite) TestCanSmartCreateWithFileStream() {
	log := Create("test", &FileStream{Path: "/var/log/test.log"})
	suite.Require().NotNil(log, "cannot create a Logger with file stream")
	suite.Assert().IsType(&FileStream{}, log.stream)
}

func (suite *InternalLoggerSuite) TestCanSmartCreateWithFileDestination() {
	log := Create("test", "/var/log/test.log")
	suite.Require().NotNil(log, "cannot create a Logger with file destination")
	suite.Assert().IsType(&FileStream{}, log.stream)
}

func (suite *InternalLoggerSuite) TestCanSmartCreateWithStreams() {
	log := Create("test", &StdoutStream{}, &StackDriverStream{})
	suite.Require().NotNil(log, "Failed to create a Logger with 2 streams")
	suite.Assert().IsType(&MultiStream{}, log.stream)
	suite.Require().Len(log.stream.(*MultiStream).streams, 2)
	suite.Assert().IsType(&StdoutStream{}, log.stream.(*MultiStream).streams[0])
	suite.Assert().IsType(&StackDriverStream{}, log.stream.(*MultiStream).streams[1])
}

func (suite *InternalLoggerSuite) TestCanSmartCreateWithDestinations() {
	log := Create("test", "stdout", "stackdriver")
	suite.Require().NotNil(log, "Failed to create a Logger with 2 destinations")
	suite.Assert().IsType(&MultiStream{}, log.stream)
	suite.Require().Len(log.stream.(*MultiStream).streams, 2)
	suite.Assert().IsType(&StdoutStream{}, log.stream.(*MultiStream).streams[0])
	suite.Assert().IsType(&StackDriverStream{}, log.stream.(*MultiStream).streams[1])
}

func (suite *InternalLoggerSuite) TestCanSmartCreateWithLogger() {
	parent := Create("test", "stdout", "stackdriver")
	log := Create("test", parent)
	suite.Require().NotNil(log, "Failed to create a Logger with parent logger")
	suite.Assert().IsType(&MultiStream{}, log.stream)
	suite.Require().Len(log.stream.(*MultiStream).streams, 2)
	suite.Assert().IsType(&StdoutStream{}, log.stream.(*MultiStream).streams[0])
	suite.Assert().IsType(&StackDriverStream{}, log.stream.(*MultiStream).streams[1])
}

func (suite *InternalLoggerSuite) TestCanSmartCreateWithRecord() {
	log := Create("test", NewRecord().Set("key1", "value1"), NewRecord().Set("key2", "value2"), NewRecord().Set("key1", "shouldnotsee"))
	suite.Require().NotNil(log, "Failed to create a Logger with a Record")
	suite.Assert().IsType(&StdoutStream{}, log.stream)
	suite.Require().NotNil(log.GetRecord("key1"), "there is no Record \"key1\" in Logger")
	suite.Assert().Equal("value1", log.GetRecord("key1").(string))
	suite.Require().NotNil(log.GetRecord("key2"), "there is no Record \"key2\" in Logger")
	suite.Assert().Equal("value2", log.GetRecord("key2").(string))
}

func (suite *InternalLoggerSuite) TestCanCreateChildWithRecords() {
	var output string
	log := Create("test", &StdoutStream{Unbuffered: true})
	suite.Require().NotNil(log, "Failed to create a Parent Logger")
	suite.Assert().IsType(&StdoutStream{}, log.stream)
	output = captureStdout(func() {
		log.Infof("Message on the Parent Logger")

		child := log.Child("childtopic", "childscope", "key1", "value1", "key2", "value2", "key1", "shouldnotsee")
		suite.Require().NotNil(child, "Failed to create a child Logger with a Record")
		child.Infof("Message on the Child Logger")

		suite.Require().NotNil(child.GetRecord("key1"), "there is no Record \"key1\" in Child Logger")
		suite.Assert().Equal("value1", child.GetRecord("key1").(string))
		suite.Require().NotNil(child.GetRecord("key2"), "there is no Record \"key2\" in Child Logger")
		suite.Assert().Equal("value2", child.GetRecord("key2").(string))

		grandchild := child.Child("grandchildtopic", "grandchildscope", "key2", "newvalue")
		suite.Require().NotNil(grandchild, "Failed to create a Grandchild Logger with a Record")
		grandchild.Infof("Message on the Grandchild Logger")
		suite.Require().NotNil(grandchild.GetRecord("key2"), "there is no Record \"key2\" in Grandchild Logger")
		suite.Assert().Equal("newvalue", grandchild.GetRecord("key2").(string))

		suite.Require().NotNil(child.GetRecord("key2"), "there is no Record \"key2\" in Child Logger")
		suite.Assert().Equal("value2", child.GetRecord("key2").(string), "Key2's value should not be changed")
	})
	suite.Assert().Nil(log.GetRecord("key1"), "there shoud not be any Record \"key1\" in Parent Logger")
	suite.Assert().Nil(log.GetRecord("key2"), "there shoud not be any Record \"key2\" in Parent Logger")
	lines := strings.Split(output, "\n")
	suite.Assert().Len(lines, 4, "Output should have 4 lines") // The 4th line is an empty line
	var content map[string]interface{}
	err := json.Unmarshal([]byte(lines[1]), &content)
	suite.Require().Nil(err, "Failed to parse JSON from the second line (child's output)")
	suite.Assert().Equal("childtopic", content["topic"], "Topic should be childtopic")
	suite.Assert().Equal("childscope", content["scope"], "Scope should be childscope")
	suite.Assert().Equal("value1", content["key1"], "key1 should contain value1")
	suite.Assert().Equal("value2", content["key2"], "key2 should contain value2")

	content = map[string]interface{}{}
	err = json.Unmarshal([]byte(lines[2]), &content)
	suite.Require().Nil(err, "Failed to parse JSON from the third (grandchild's output)")
	suite.Assert().Equal("grandchildtopic", content["topic"], "Topic should be grandchildtopic")
	suite.Assert().Equal("grandchildscope", content["scope"], "Scope should be grandchildscope")
	suite.Assert().Equal("value1", content["key1"], "key1 should contain value1")
	suite.Assert().Equal("newvalue", content["key2"], "key2 should contain newvalue")
}

func (suite *InternalLoggerSuite) TestCanSmartCreateWithMix() {
	log := Create("test", &StderrStream{}, "/var/log/test.log", NewRecord().Set("key", "value"))
	suite.Require().NotNil(log, "Failed to create a Logger with a Record")
	suite.Assert().IsType(&MultiStream{}, log.stream)
	suite.Require().Len(log.stream.(*MultiStream).streams, 2)
	suite.Assert().IsType(&StderrStream{}, log.stream.(*MultiStream).streams[0])
	suite.Assert().IsType(&FileStream{}, log.stream.(*MultiStream).streams[1])
	suite.Assert().Equal("/var/log/test.log", log.stream.(*MultiStream).streams[1].(*FileStream).Path)
	suite.Require().NotNil(log.GetRecord("key"), "there is no Record \"key\" in Logger")
	suite.Assert().Equal("value", log.GetRecord("key").(string))
}

func (suite *InternalLoggerSuite) TestCanSetFilterLevel() {
	log := Create("test")
	suite.Require().NotNil(log, "cannot create a Logger")
	suite.Assert().IsType(&StdoutStream{}, log.stream)
	suite.Assert().Equal(INFO, log.stream.(*StdoutStream).FilterLevel, "FilterLevel should be INFO")
	log.SetFilterLevel(WARN)
	suite.Assert().Equal(WARN, log.stream.(*StdoutStream).FilterLevel, "FilterLevel should be WARN")
	log.SetFilterLevelIfUnset(ERROR)
	suite.Assert().Equal(WARN, log.stream.(*StdoutStream).FilterLevel, "FilterLevel should be WARN")
}

func(suite *InternalLoggerSuite) TestCanConvertBytesToString() {
	suite.Assert().Equal("12B", bytesToString(uint64(12)))
	suite.Assert().Equal("12.00KiB", bytesToString(uint64(12 * 1024)))
	suite.Assert().Equal("12.00MiB", bytesToString(uint64(12 * 1024 * 1024)))
	suite.Assert().Equal("12.00GiB", bytesToString(uint64(12 * 1024 * 1024 * 1024)))
}

func captureStdout(f func()) string {
	reader, writer, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	stdout := os.Stdout
	os.Stdout = writer
	defer func() {
		os.Stdout = stdout
	}()

	f()
	writer.Close()

	output := bytes.Buffer{}
	_, _ = io.Copy(&output, reader)
	return output.String()
}
