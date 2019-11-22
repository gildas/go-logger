package logger_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

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

	stream = logger.CreateStreamWithDestination("null")
	suite.Require().NotNil(stream, "Failed to create a nil stream")
	suite.Assert().IsType(&logger.NilStream{}, stream)

	stream = logger.CreateStreamWithDestination("void")
	suite.Require().NotNil(stream, "Failed to create a nil stream")
	suite.Assert().IsType(&logger.NilStream{}, stream)

	stream = logger.CreateStreamWithDestination("gcp")
	suite.Require().NotNil(stream, "Failed to create a nil stream")
	suite.Assert().IsType(&logger.GCPStream{}, stream)

	stream = logger.CreateStreamWithDestination("google")
	suite.Require().NotNil(stream, "Failed to create a nil stream")
	suite.Assert().IsType(&logger.GCPStream{}, stream)

	stream = logger.CreateStreamWithDestination("stackdriver")
	suite.Require().NotNil(stream, "Failed to create a nil stream")
	suite.Assert().IsType(&logger.StackDriverStream{}, stream)

	stream = logger.CreateStreamWithDestination("file://./log/test.log")
	suite.Require().NotNil(stream, "Failed to create a nil stream")
	suite.Assert().IsType(&logger.FileStream{}, stream)

	stream = logger.CreateStreamWithDestination("/var/log/test.log")
	suite.Require().NotNil(stream, "Failed to create a nil stream")
	suite.Assert().IsType(&logger.FileStream{}, stream)

	stream = logger.CreateStreamWithDestination("./log/test.log")
	suite.Require().NotNil(stream, "Failed to create a nil stream")
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

func (suite *StreamSuite) TestCanStreamToFile() {
	folder, teardown := CreateTempDir(suite.T())
	defer teardown()
	stream := &logger.FileStream{Path: filepath.Join(folder, "test.log")}

	record := logger.NewRecord().Set("bello", "banana").Set("だれ", "Me")
	err := stream.Write(record)
	suite.Require().Nil(err)
	stream.Flush()

	payload, err := json.Marshal(record)
	suite.Require().Nil(err, "Failed to marshar Record")

	content, err := ioutil.ReadFile(stream.Path)
	suite.Require().Nil(err, "Failed to read %s", stream.Path)
	suite.Assert().JSONEq(string(payload), string(content))
}

func ExampleStdoutStream() {
	stream := &logger.StdoutStream{}
	record := logger.NewRecord().Set("bello", "banana").Set("だれ", "Me")

	err := stream.Write(record)
	if err != nil {
		os.Stdout.WriteString(err.Error() + "\n")
	}
	stream.Flush()
	// Output: {"bello":"banana","だれ":"Me"}
}

func (suite *StreamSuite) TestCanCreateFileStream() {
}

func (suite *StreamSuite) SetupSuite() {
	suite.Name = strings.TrimSuffix(reflect.TypeOf(*suite).Name(), "Suite")
}