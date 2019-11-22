package logger

import (
	"testing"

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