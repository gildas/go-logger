package logger

import (
	"bytes"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type InternalLoggerSuite struct {
	suite.Suite
	Name string
}

func TestInternalLoggerSuite(t *testing.T) {
	suite.Run(t, new(InternalLoggerSuite))
}

func (suite *InternalLoggerSuite) TestCanCreateWithStreams() {
	var log *Logger

	log = CreateWithStream("test")
	suite.Require().NotNil(log, "Failed to create a Logger with no stream")
	suite.Assert().IsType(&StdoutStream{}, log.stream)

	log = CreateWithStream("test", &NilStream{})
	suite.Require().NotNil(log, "Failed to create a Logger with a nil stream")
	suite.Assert().IsType(&NilStream{}, log.stream)

	log = CreateWithStream("test", &StdoutStream{}, &StackDriverStream{})
	suite.Require().NotNil(log, "Failed to create a Logger with 2 streams")
	suite.Assert().IsType(&MultiStream{}, log.stream)
	suite.Require().Len(log.stream.(*MultiStream).streams, 2)
	suite.Assert().IsType(&StdoutStream{}, log.stream.(*MultiStream).streams[0])
	suite.Assert().IsType(&StackDriverStream{}, log.stream.(*MultiStream).streams[1])

	os.Setenv("DEBUG", "1")
	log = CreateWithStream("test")
	suite.Require().NotNil(log, "Failed to create a Logger with stdout stream")
	suite.Assert().IsType(&StdoutStream{}, log.stream)
	suite.Assert().Equal(true, log.stream.(*StdoutStream).Unbuffered, "In DEBUG mode, stdout should be unbuffered")
	os.Unsetenv("DEBUG")

	os.Setenv("LOG_FLUSHFREQUENCY", "10ms")
	log = CreateWithStream("test", &StdoutStream{})
	suite.Require().NotNil(log, "Failed to create a Logger with stdout stream")
	suite.Assert().IsType(&StdoutStream{}, log.stream)
	_ = captureStdout(func() {
		log.Tracef("writing something")
	})
	suite.Assert().Equal(10*time.Millisecond, log.stream.(*StdoutStream).flushFrequency, "this stream should flush every 10 milliseconds")
	os.Unsetenv("LOG_FLUSHFREQUENCY")
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
