package logger_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/gildas/go-logger"
)

type StreamSuite struct {
	suite.Suite
	Name  string
}

func TestStreamSuite(t *testing.T) {
	suite.Run(t, new(StreamSuite))
}

func (suite *StreamSuite) TestCanCreateStreamFromDestination() {
	var stream logger.Streamer

	stream = logger.CreateStreamWithDestination("nil")
	suite.Require().NotNil(stream, "Failed to create a nil stream")
	suite.Assert().IsType(&logger.NilStream{}, stream)
	suite.Assert().IsType(&logger.NilStream{}, stream)
	suite.Assert().Equal("Stream to nil", fmt.Sprintf("%s", stream))

	stream = logger.CreateStreamWithDestination("null")
	suite.Require().NotNil(stream, "Failed to create a nil stream")
	suite.Assert().IsType(&logger.NilStream{}, stream)
	suite.Assert().IsType(&logger.NilStream{}, stream)
	suite.Assert().Equal("Stream to nil", fmt.Sprintf("%s", stream))

	stream = logger.CreateStreamWithDestination("void")
	suite.Require().NotNil(stream, "Failed to create a nil stream")
	suite.Assert().IsType(&logger.NilStream{}, stream)
	suite.Assert().Equal("Stream to nil", fmt.Sprintf("%s", stream))

	stream = logger.CreateStreamWithDestination("stdout")
	suite.Require().NotNil(stream, "Failed to create a stdout stream")
	suite.Assert().IsType(&logger.StdoutStream{}, stream)
	suite.Assert().Equal("Stream to stdout", fmt.Sprintf("%s", stream))

	stream = logger.CreateStreamWithDestination("stderr")
	suite.Require().NotNil(stream, "Failed to create a stderr stream")
	suite.Assert().IsType(&logger.StderrStream{}, stream)
	suite.Assert().Equal("Stream to stderr", fmt.Sprintf("%s", stream))

	stream = logger.CreateStreamWithDestination("gcp")
	suite.Require().NotNil(stream, "Failed to create a Google Cloud Platform stream")
	suite.Assert().IsType(&logger.GCPStream{}, stream)

	stream = logger.CreateStreamWithDestination("google")
	suite.Require().NotNil(stream, "Failed to create a Google Cloud Platform stream")
	suite.Assert().IsType(&logger.GCPStream{}, stream)

	stream = logger.CreateStreamWithDestination("stackdriver")
	suite.Require().NotNil(stream, "Failed to create a Google Stackdriver stream")
	suite.Assert().IsType(&logger.StackDriverStream{}, stream)

	stream = logger.CreateStreamWithDestination("file://./log/test.log")
	suite.Require().NotNil(stream, "Failed to create a file stream")
	suite.Assert().IsType(&logger.FileStream{}, stream)

	stream = logger.CreateStreamWithDestination("/var/log/test.log")
	suite.Require().NotNil(stream, "Failed to create a file stream")
	suite.Assert().IsType(&logger.FileStream{}, stream)

	stream = logger.CreateStreamWithDestination("./log/test.log")
	suite.Require().NotNil(stream, "Failed to create a file stream")
	suite.Assert().IsType(&logger.FileStream{}, stream)

	stream = logger.CreateStreamWithDestination()
	suite.Require().NotNil(stream, "Failed to create a stream from an empty destination")
	suite.Assert().IsType(&logger.StdoutStream{}, stream)

	stream = logger.CreateStreamWithDestination("")
	suite.Require().NotNil(stream, "Failed to create a stream from an empty destination")
	suite.Assert().IsType(&logger.StdoutStream{}, stream)

	stream = logger.CreateStreamWithDestination("myfile", "stackdriver")
	suite.Require().NotNil(stream, "Failed to create a stream from an empty destination")
	suite.Assert().IsType(&logger.MultiStream{}, stream)
}

func (suite *StreamSuite) TestCanCreateStdoutStream() {
	stream := &logger.StdoutStream{}
	suite.Assert().Equal("Stream to stdout", stream.String())
}

func (suite *StreamSuite) TestCanCreateUnbufferedStdoutStream() {
	stream := &logger.StdoutStream{Unbuffered: true, FilterLevel: logger.INFO}
	suite.Assert().Equal("Unbuffered Stream to stdout, Filter: INFO", stream.String())
}

func (suite *StreamSuite) TestCanCreateFileStream() {
	stream := &logger.FileStream{Path: "/tmp/test.log"}
	suite.Assert().Equal("Stream to /tmp/test.log", stream.String())
}

func (suite *StreamSuite) TestCanCreateUnbufferedFileStream() {
	stream := &logger.FileStream{Path: "/tmp/test.log", Unbuffered: true, FilterLevel: logger.INFO}
	suite.Assert().Equal("Unbuffered Stream to /tmp/test.log, Filter: INFO", stream.String())
}

func (suite *StreamSuite) TestCanStreamToFile() {
	os.Setenv("LOG_FLUSHFREQUENCY", "10ms")
	defer os.Unsetenv("LOG_FLUSHFREQUENCY")
	folder, teardown := CreateTempDir(suite.T())
	defer teardown()
	stream := &logger.FileStream{Path: filepath.Join(folder, "test.log")}

	record := logger.NewRecord().Set("bello", "banana").Set("だれ", "Me")
	err := stream.Write(record)
	suite.Require().Nil(err)
	time.Sleep(11 * time.Millisecond)
	stream.Flush()

	payload, err := json.Marshal(record)
	suite.Require().Nil(err, "Failed to marshar Record")

	content, err := ioutil.ReadFile(stream.Path)
	suite.Require().Nil(err, "Failed to read %s", stream.Path)
	suite.Assert().JSONEq(string(payload), string(content))
}

func ExampleStdoutStream() {
	os.Setenv("LOG_FLUSHFREQUENCY", "10ms")
	defer os.Unsetenv("LOG_FLUSHFREQUENCY")
	stream := &logger.StdoutStream{}

	if err := stream.Write(logger.NewRecord().Set("bello", "banana").Set("だれ", "Me")); err != nil {
		os.Stdout.WriteString(err.Error() + "\n")
	}
	if stream.ShouldWrite(logger.TRACE) {
		os.Stdout.WriteString("This should not be seen, stream Filter: " + stream.FilterLevel.String() + "\n")
	}
	if err := stream.Write(logger.NewRecord().Set("bello", "banana").Set("level", logger.ERROR)); err != nil {
		os.Stdout.WriteString(err.Error() + "\n")
	}
	stream.Flush()
	time.Sleep(11 * time.Millisecond)
	// Output: 
	// {"bello":"banana","だれ":"Me"}
	// {"bello":"banana","level":50}
}

func ExampleStdoutStream_Unbuffered() {
	stream := &logger.StdoutStream{ Unbuffered: true}

	if err := stream.Write(logger.NewRecord().Set("bello", "banana").Set("だれ", "Me")); err != nil {
		os.Stdout.WriteString(err.Error() + "\n")
	}
	if stream.ShouldWrite(logger.TRACE) {
		os.Stdout.WriteString("This should not be seen, stream Filter: " + stream.FilterLevel.String() + "\n")
	}
	// Output: {"bello":"banana","だれ":"Me"}
}

func ExampleStderrStream() {
	output := CaptureStderr(func() {
		stream := &logger.StderrStream{}

		if err := stream.Write(logger.NewRecord().Set("bello", "banana").Set("だれ", "Me")); err != nil {
			os.Stderr.WriteString(err.Error() + "\n")
		}
		if stream.ShouldWrite(logger.TRACE) {
			os.Stderr.WriteString("This should not be seen, stream Filter: " + stream.FilterLevel.String() + "\n")
		}
        stream.Flush()
	})
	fmt.Println(output)
	// Output: {"bello":"banana","だれ":"Me"}
}

func ExampleNilStream() {
	stream := &logger.NilStream{}

	if err := stream.Write(logger.NewRecord().Set("bello", "banana").Set("だれ", "Me")); err != nil {
		os.Stdout.WriteString(err.Error() + "\n")
	}
	if stream.ShouldWrite(logger.ALWAYS) {
		os.Stdout.WriteString("This should not be seen\n")
	}
	stream.Flush()
	// Output:
}

func (suite *StreamSuite) TestCanStreamToGCP() {
	_ = CaptureStdout(func() {
		stream := &logger.GCPStream{}
		suite.Assert().Equal("Stream to Google Cloud", stream.String())
		suite.Assert().Truef(stream.ShouldWrite(logger.WARN), "It should be possible to write to a %s", stream)
		err := stream.Write(logger.NewRecord().Set("bello", "banana").Set("level", logger.NEVER))
		suite.Assert().Nil(err, "Failed to write to stream")
		err = stream.Write(logger.NewRecord().Set("bello", "banana").Set("level", logger.TRACE))
		suite.Assert().Nil(err, "Failed to write to stream")
		err = stream.Write(logger.NewRecord().Set("bello", "banana").Set("level", logger.DEBUG))
		suite.Assert().Nil(err, "Failed to write to stream")
		err = stream.Write(logger.NewRecord().Set("bello", "banana").Set("level", logger.INFO))
		suite.Assert().Nil(err, "Failed to write to stream")
		err = stream.Write(logger.NewRecord().Set("bello", "banana").Set("level", logger.WARN))
		suite.Assert().Nil(err, "Failed to write to stream")
		err = stream.Write(logger.NewRecord().Set("bello", "banana").Set("level", logger.ERROR))
		suite.Assert().Nil(err, "Failed to write to stream")
		err = stream.Write(logger.NewRecord().Set("bello", "banana").Set("level", logger.FATAL))
		suite.Assert().Nil(err, "Failed to write to stream")
		err = stream.Write(logger.NewRecord().Set("bello", "banana").Set("level", logger.ALWAYS))
		suite.Assert().Nil(err, "Failed to write to stream")
		err = stream.Write(logger.NewRecord().Set("bello", "banana").Set("level", logger.UNSET))
		suite.Assert().Nil(err, "Failed to write to stream")
		stream.Flush()
	})
}

func (suite *StreamSuite) TestCanStreamToStackDriver() {
	stream := &logger.StackDriverStream{}
	suite.Assert().Equal("Stream to Google StackDriver", stream.String())
	suite.Assert().Truef(stream.ShouldWrite(logger.WARN), "It should be possible to write to a %s", stream)
	stream.Flush()
}

func (suite *StreamSuite) TestCanCreateMultiStream() {
	stream := logger.CreateStreamWithDestination("stdout", "nil")
	suite.Assert().IsType(&logger.MultiStream{}, stream)
	suite.Assert().Equal("MultiStream", fmt.Sprintf("%s", stream))
	suite.Assert().Truef(stream.ShouldWrite(logger.WARN), "It should be possible to write to a %s", stream)
	err := stream.Write(logger.NewRecord().Set("bello", "banana").Set("level", logger.ERROR))
	suite.Assert().Nil(err, "Failed to write to stream")
	stream.Flush()
}

func (suite *StreamSuite) TestCanGetFlushFrequencyFromEnvironment() {
	os.Unsetenv("LOG_FLUSHFREQUENCY")
	frequency := logger.GetFlushFrequencyFromEnvironment()
	suite.Assert().Equal(5 * time.Minute, frequency, "Frequency should be 5 minutes before being set in the environment")
	os.Setenv("LOG_FLUSHFREQUENCY", "3600s")
	frequency = logger.GetFlushFrequencyFromEnvironment()
	suite.Assert().Equal(1 * time.Hour, frequency, "Frequency should be 1 hour after being set in the environment (was %s)", frequency)
	os.Setenv("LOG_FLUSHFREQUENCY", "P2H")
	frequency = logger.GetFlushFrequencyFromEnvironment()
	suite.Assert().Equal(2 * time.Hour, frequency, "Frequency should be 2 hour after being set in the environment (was %s)", frequency)
	os.Unsetenv("LOG_FLUSHFREQUENCY")
}

func (suite *StreamSuite) TestCanCreateStreamFromEnvironment() {
	os.Setenv("LOG_DESTINATION", "/var/log/test.log")
	stream := logger.CreateStreamWithDestination()
	suite.Require().NotNil(stream, "Failed to create a file stream")
	suite.Assert().IsType(&logger.FileStream{}, stream)
	suite.Assert().Equal("/var/log/test.log", stream.(*logger.FileStream).Path, "File Stream Path should be /var/log/test.log")
	os.Unsetenv("LOG_DESTINATION")
}

func (suite *StreamSuite) TestFileStreamFailsWritingToInvalidFile() {
	streamFile := &logger.FileStream{Path: ""}
	err := streamFile.Write(logger.NewRecord().Set("key", "value"))
	suite.Assert().NotNil(err)
	suite.Assert().Contains(err.Error(), "no such file")
}

func (suite *StreamSuite) TestFailsWritingtoMultiStreamWithBogusStream() {
	stream := logger.CreateMultiStream(&logger.StdoutStream{}, &BogusStream{})
	suite.Assert().IsType(&logger.MultiStream{}, stream)
	err := stream.Write(logger.NewRecord().Set("bello", "banana").Set("だれ", "Me"))
	suite.Assert().NotNil(err, "Should have failed writing to stream")

}

func (suite *StreamSuite) TestFailsWritingWithBogusRecordValue() {
	streamStderr := &logger.StderrStream{}
	err := streamStderr.Write(logger.NewRecord().Set("key", &BogusValue{}))
	suite.Assert().NotNil(err)
	suite.Assert().Contains(err.Error(), "Failed to Marshal BogusValue")
	streamStdout := &logger.StdoutStream{}
	err = streamStdout.Write(logger.NewRecord().Set("key", &BogusValue{}))
	suite.Assert().NotNil(err)
	suite.Assert().Contains(err.Error(), "Failed to Marshal BogusValue")
	streamFile := &logger.FileStream{Path: "/tmp/test.log"}
	err = streamFile.Write(logger.NewRecord().Set("key", &BogusValue{}))
	suite.Assert().NotNil(err)
	suite.Assert().Contains(err.Error(), "Failed to Marshal BogusValue")
	streamGCP := &logger.GCPStream{}
	err = streamGCP.Write(logger.NewRecord().Set("key", &BogusValue{}).Set("level", logger.INFO))
	suite.Assert().NotNil(err)
	suite.Assert().Contains(err.Error(), "Failed to Marshal BogusValue")
}

func (suite *StreamSuite) SetupSuite() {
	suite.Name = strings.TrimSuffix(reflect.TypeOf(*suite).Name(), "Suite")
}