package logger_test

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/gildas/go-errors"
	"github.com/gildas/go-logger"
)

type StreamSuite struct {
	suite.Suite
	Name string
}

func TestStreamSuite(t *testing.T) {
	suite.Run(t, new(StreamSuite))
}

func (suite *StreamSuite) SetupSuite() {
	suite.Name = strings.TrimSuffix(reflect.TypeOf(suite).Elem().Name(), "Suite")
}

func (suite *StreamSuite) TestCanCreateStreamFromDestination() {
	var stream logger.Streamer

	stream = logger.CreateStream(logger.NewLevelSet(logger.INFO), "nil")
	suite.Require().NotNil(stream, "Failed to create a nil stream")
	suite.Assert().IsType(&logger.NilStream{}, stream)
	suite.Assert().IsType(&logger.NilStream{}, stream)
	suite.Assert().Equal("Stream to nil", fmt.Sprintf("%s", stream))
	suite.Assert().False(stream.ShouldLogSourceInfo(), "Should not log source info")
	stream.Close()

	stream = logger.CreateStream(logger.NewLevelSet(logger.INFO), "null")
	suite.Require().NotNil(stream, "Failed to create a nil stream")
	suite.Assert().IsType(&logger.NilStream{}, stream)
	suite.Assert().IsType(&logger.NilStream{}, stream)
	suite.Assert().Equal("Stream to nil", fmt.Sprintf("%s", stream))
	suite.Assert().False(stream.ShouldLogSourceInfo(), "Should not log source info")
	suite.Assert().Equal(logger.LevelSet{}, stream.GetFilterLevels())
	stream.Close()

	stream = logger.CreateStream(logger.NewLevelSet(logger.INFO), "void")
	suite.Require().NotNil(stream, "Failed to create a nil stream")
	suite.Assert().IsType(&logger.NilStream{}, stream)
	suite.Assert().Equal("Stream to nil", fmt.Sprintf("%s", stream))
	suite.Assert().False(stream.ShouldLogSourceInfo(), "Should not log source info")
	suite.Assert().Equal(logger.LevelSet{}, stream.GetFilterLevels())
	stream.Close()

	stream = logger.CreateStream(logger.NewLevelSet(logger.INFO), "stdout")
	suite.Require().NotNil(stream, "Failed to create a stdout stream")
	suite.Assert().IsType(&logger.StdoutStream{}, stream)
	suite.Assert().Equal("Stream to stdout, Filter: INFO", fmt.Sprintf("%s", stream))
	suite.Assert().False(stream.ShouldLogSourceInfo(), "Should not log source info")
	suite.Assert().Equal(logger.NewLevelSet(logger.INFO), stream.GetFilterLevels())
	stream.Close()

	stream = logger.CreateStream(logger.NewLevelSet(logger.INFO), "stderr")
	suite.Require().NotNil(stream, "Failed to create a stderr stream")
	suite.Assert().IsType(&logger.StderrStream{}, stream)
	suite.Assert().Equal("Stream to stderr, Filter: INFO", fmt.Sprintf("%s", stream))
	suite.Assert().False(stream.ShouldLogSourceInfo(), "Should not log source info")
	suite.Assert().Equal(logger.NewLevelSet(logger.INFO), stream.GetFilterLevels())
	stream.Close()

	stream = logger.CreateStream(logger.NewLevelSet(logger.INFO), "gcp")
	suite.Require().NotNil(stream, "Failed to create a Google Cloud Platform stream")
	suite.Assert().IsType(&logger.StdoutStream{}, stream)
	suite.Assert().NotNil((stream.(*logger.StdoutStream)).Converter)
	suite.Assert().IsType(&logger.StackDriverConverter{}, (stream.(*logger.StdoutStream)).Converter)
	suite.Assert().False(stream.ShouldLogSourceInfo(), "Should not log source info")
	suite.Assert().Equal(logger.NewLevelSet(logger.INFO), stream.GetFilterLevels())
	stream.Close()

	stream = logger.CreateStream(logger.NewLevelSet(logger.INFO), "google")
	suite.Require().NotNil(stream, "Failed to create a Google Cloud Platform stream")
	suite.Assert().IsType(&logger.StdoutStream{}, stream)
	suite.Assert().False(stream.ShouldLogSourceInfo(), "Should not log source info")
	suite.Assert().Equal(logger.NewLevelSet(logger.INFO), stream.GetFilterLevels())
	stream.Close()

	stream = logger.CreateStream(logger.NewLevelSet(logger.INFO), "stackdriver")
	suite.Require().NotNil(stream, "Failed to create a Google Stackdriver stream")
	suite.Assert().IsType(&logger.StackDriverStream{}, stream)
	suite.Assert().Equal("Stream to Google StackDriver, Filter: INFO", fmt.Sprintf("%s", stream))
	suite.Assert().False(stream.ShouldLogSourceInfo(), "Should not log source info")
	suite.Assert().Equal(logger.NewLevelSet(logger.INFO), stream.GetFilterLevels())
	stream.Close()

	stream = logger.CreateStream(logger.NewLevelSet(logger.INFO), "file://./log/test.log")
	suite.Require().NotNil(stream, "Failed to create a file stream")
	suite.Assert().IsType(&logger.FileStream{}, stream)
	suite.Assert().False(stream.ShouldLogSourceInfo(), "Should not log source info")
	suite.Assert().Equal(logger.NewLevelSet(logger.INFO), stream.GetFilterLevels())
	stream.Close()

	stream = logger.CreateStream(logger.NewLevelSet(logger.INFO), "/var/log/test.log")
	suite.Require().NotNil(stream, "Failed to create a file stream")
	suite.Assert().IsType(&logger.FileStream{}, stream)
	suite.Assert().False(stream.ShouldLogSourceInfo(), "Should not log source info")
	suite.Assert().Equal(logger.NewLevelSet(logger.INFO), stream.GetFilterLevels())
	stream.Close()

	stream = logger.CreateStream(logger.NewLevelSet(logger.INFO), "./log/test.log")
	suite.Require().NotNil(stream, "Failed to create a file stream")
	suite.Assert().IsType(&logger.FileStream{}, stream)
	suite.Assert().False(stream.ShouldLogSourceInfo(), "Should not log source info")
	suite.Assert().Equal(logger.NewLevelSet(logger.INFO), stream.GetFilterLevels())
	stream.Close()

	stream = logger.CreateStream(logger.NewLevelSet(logger.INFO))
	suite.Require().NotNil(stream, "Failed to create a stream from an empty destination")
	suite.Assert().IsType(&logger.StdoutStream{}, stream)
	suite.Assert().False(stream.ShouldLogSourceInfo(), "Should not log source info")
	suite.Assert().Equal(logger.NewLevelSet(logger.INFO), stream.GetFilterLevels())
	stream.Close()

	stream = logger.CreateStream(logger.NewLevelSet(logger.INFO), "")
	suite.Require().NotNil(stream, "Failed to create a stream from an empty destination")
	suite.Assert().IsType(&logger.StdoutStream{}, stream)
	suite.Assert().False(stream.ShouldLogSourceInfo(), "Should not log source info")
	suite.Assert().Equal(logger.NewLevelSet(logger.INFO), stream.GetFilterLevels())
	stream.Close()

	stream = logger.CreateStream(logger.NewLevelSet(logger.INFO), "myfile", "stackdriver")
	suite.Require().NotNil(stream, "Failed to create a stream from an empty destination")
	suite.Assert().IsType(&logger.MultiStream{}, stream)
	suite.Assert().False(stream.ShouldLogSourceInfo(), "Should not log source info")
	suite.Assert().Equal(logger.NewLevelSet(logger.INFO), stream.GetFilterLevels())
	suite.Assert().Equal(logger.LevelSet{}, (&logger.MultiStream{}).GetFilterLevels())
	stream.Close()
}

func (suite *StreamSuite) TestCanCreateStreamFromEnvironment() {
	_ = os.Setenv("LOG_DESTINATION", "/var/log/test.log")
	defer func() { _ = os.Unsetenv("LOG_DESTINATION") }()
	stream := logger.CreateStream(logger.NewLevelSet(logger.INFO))
	suite.Require().NotNil(stream, "Failed to create a file stream")
	suite.Assert().IsType(&logger.FileStream{}, stream)
	suite.Assert().Equal("/var/log/test.log", stream.(*logger.FileStream).Path, "File Stream Path should be /var/log/test.log")
}

func (suite *StreamSuite) TestCanCreateStderrStream() {
	stream := &logger.StderrStream{}
	suite.Assert().Equal("Stream to stderr", stream.String())
	stream.Flush()
}

func (suite *StreamSuite) TestCanCreateStdoutStream() {
	stream := &logger.StdoutStream{}
	suite.Assert().Equal("Stream to stdout", stream.String())
	stream.Flush()
}

func (suite *StreamSuite) TestCanCreateUnbufferedStdoutStream() {
	stream := &logger.StdoutStream{Unbuffered: true, FilterLevels: logger.NewLevelSet(logger.INFO)}
	suite.Assert().Equal("Unbuffered Stream to stdout, Filter: INFO", stream.String())
	stream.Flush()
}

func (suite *StreamSuite) TestCanCreateFileStream() {
	stream := &logger.FileStream{Path: "log/test.log"}
	suite.Assert().Equal("Stream to log/test.log", stream.String())
	err := stream.Write(logger.NewRecord().Set("key", "value"))
	suite.Assert().Nil(err, "FileStream should have written something")
	stream.Flush()
}

func (suite *StreamSuite) TestCanCreateUnbufferedFileStream() {
	stream := &logger.FileStream{Path: "log/test.log", Unbuffered: true, FilterLevels: logger.NewLevelSet(logger.INFO)}
	suite.Assert().Equal("Unbuffered Stream to log/test.log, Filter: INFO", stream.String())
	err := stream.Write(logger.NewRecord().Set("key", "value"))
	suite.Assert().Nil(err, "FileStream should have written something")
}

func (suite *StreamSuite) TestCanCreateStackDriverStream() {
	stream := &logger.StackDriverStream{Parent: "go-logger-test", KeyFilename: "gcloud-key.json"}
	suite.Assert().Equal("Stream to Google StackDriver", stream.String())
}

func (suite *StreamSuite) TestCanCreateMultiStream() {
	stream := logger.CreateMultiStream(&logger.StdoutStream{}, &logger.StderrStream{SourceInfo: true})
	suite.Assert().IsType(&logger.MultiStream{}, stream)
	suite.Assert().Equal("MultiStream", fmt.Sprintf("%s", stream))
	suite.Assert().True(stream.ShouldLogSourceInfo(), "Should log source info")
}

func (suite *StreamSuite) TestFileStreamCanSetFilterLevel() {
	stream := &logger.FileStream{}
	suite.Assert().Equal(logger.UNSET, stream.FilterLevels.GetDefault())
	stream.SetFilterLevel(logger.WARN)
	suite.Assert().Equal(logger.WARN, stream.FilterLevels.GetDefault())
}

func (suite *StreamSuite) TestNilStreamCanSetFilterLevel() {
	stream := &logger.NilStream{}
	stream.SetFilterLevel(logger.WARN)
}

func (suite *StreamSuite) TestStderrStreamCanSetFilterLevel() {
	stream := &logger.StderrStream{}
	suite.Assert().Equal(logger.UNSET, stream.FilterLevels.GetDefault())
	stream.SetFilterLevel(logger.WARN)
	suite.Assert().Equal(logger.WARN, stream.FilterLevels.GetDefault())
}

func (suite *StreamSuite) TestStdoutStreamCanSetFilterLevel() {
	stream := &logger.StdoutStream{}
	suite.Assert().Equal(logger.UNSET, stream.FilterLevels.GetDefault())
	stream.SetFilterLevel(logger.WARN)
	suite.Assert().Equal(logger.WARN, stream.FilterLevels.GetDefault())
}

func (suite *StreamSuite) TestStackDriverStreamCanSetFilterLevel() {
	stream := &logger.StackDriverStream{}
	suite.Assert().Equal(logger.UNSET, stream.FilterLevels.GetDefault())
	stream.SetFilterLevel(logger.WARN)
	suite.Assert().Equal(logger.WARN, stream.FilterLevels.GetDefault())
}

func (suite *StreamSuite) TestCanStreamToFile() {
	var err error
	_ = os.Setenv("LOG_FLUSHFREQUENCY", "10ms")
	defer func() { _ = os.Unsetenv("LOG_FLUSHFREQUENCY") }()
	folder, teardown := CreateTempDir()
	defer teardown()
	stream := &logger.FileStream{Path: filepath.Join(folder, "test.log")}
	defer stream.Close()

	err = stream.Write(logger.NewRecord().Set("bello", "banana").Set("だれ", "私"))
	suite.Require().NoError(err)
	err = stream.Write(logger.NewRecord().Set("key1", "value1").Set("key2", "value2"))
	suite.Require().NoError(err)
	time.Sleep(11 * time.Millisecond)
	stream.Flush()

	content, err := os.ReadFile(stream.Path)
	suite.Require().NoError(err, "Failed to read %s", stream.Path)
	lines := strings.Split(string(content), "\n")
	suite.Require().Len(lines, 3, "Should have written 2 lines")
	suite.Assert().JSONEq(string(`{"bello":"banana","だれ":"私"}`), lines[0])
	suite.Assert().JSONEq(string(`{"key1":"value1","key2":"value2"}`), lines[1])
}

func (suite *StreamSuite) TestCanStreamToUnbufferedFile() {
	var err error
	folder, teardown := CreateTempDir()
	defer teardown()
	stream := &logger.FileStream{Path: filepath.Join(folder, "test.log"), Unbuffered: true}
	defer stream.Close()

	err = stream.Write(logger.NewRecord().Set("bello", "banana").Set("だれ", "私"))
	suite.Require().NoError(err)
	err = stream.Write(logger.NewRecord().Set("key1", "value1").Set("key2", "value2"))
	suite.Require().NoError(err)

	content, err := os.ReadFile(stream.Path)
	suite.Require().NoError(err, "Failed to read %s", stream.Path)
	lines := strings.Split(string(content), "\n")
	suite.Require().Len(lines, 3, "Should have written 2 lines")
	suite.Assert().JSONEq(string(`{"bello":"banana","だれ":"私"}`), lines[0])
	suite.Assert().JSONEq(string(`{"key1":"value1","key2":"value2"}`), lines[1])
}

func (suite *StreamSuite) TestCanStreamToStdout() {
	_ = os.Setenv("LOG_FLUSHFREQUENCY", "10ms")
	defer func() { _ = os.Unsetenv("LOG_FLUSHFREQUENCY") }()
	stream := &logger.StdoutStream{}
	output := CaptureStdout(func() {
		err := stream.Write(logger.NewRecord().Set("bello", "banana").Set("だれ", "私"))
		suite.Require().Nil(err)
		stream.Flush()
		time.Sleep(11 * time.Millisecond)
		stream.Close()
	})
	lines := strings.Split(output, "\n")
	suite.Require().Len(lines, 2, "Should have written 1 line")
	suite.Assert().JSONEq(string(`{"bello":"banana","だれ":"私"}`), lines[0])
}

func (suite *StreamSuite) TestCanStreamToUnbufferedStdout() {
	stream := &logger.StdoutStream{Unbuffered: true}
	output := CaptureStdout(func() {
		err := stream.Write(logger.NewRecord().Set("bello", "banana").Set("だれ", "私"))
		suite.Require().Nil(err)
	})
	lines := strings.Split(output, "\n")
	suite.Require().Len(lines, 2, "Should have written 1 line")
	suite.Assert().JSONEq(string(`{"bello":"banana","だれ":"私"}`), lines[0])
}

func (suite *StreamSuite) TestCanStreamToStderr() {
	stream := &logger.StderrStream{}
	output := CaptureStderr(func() {
		err := stream.Write(logger.NewRecord().Set("bello", "banana").Set("だれ", "私"))
		suite.Require().Nil(err)
	})
	lines := strings.Split(output, "\n")
	suite.Require().Len(lines, 2, "Should have written 1 line")
	suite.Assert().JSONEq(string(`{"bello":"banana","だれ":"私"}`), lines[0])
}

func (suite *StreamSuite) TestCanStreamToNil() {
	stream := &logger.NilStream{}
	output := CaptureStdout(func() {
		err := stream.Write(logger.NewRecord().Set("bello", "banana").Set("だれ", "私"))
		suite.Require().Nil(err)
	})
	suite.Assert().Empty(output, "Should not have written anything")
}

func (suite *StreamSuite) TestCanStreamToStackDriver() {
	if _, ok := os.LookupEnv("GOOGLE_PROJECT_ID"); !ok {
		suite.T().Skip("There is no way to test this without a Google Project ID")
	}
	stream := &logger.StackDriverStream{LogID: "test"}
	defer stream.Close()
	suite.Assert().Equal("Stream to Google StackDriver", stream.String())
	suite.Assert().Truef(stream.ShouldWrite(logger.WARN, "", ""), "It should be possible to write to a %s", stream)
	if _, ok := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS"); ok {
		err := stream.Write(logger.NewRecord().Set("bello", "banana").Set("level", logger.WARN).Set("time", time.Now()).Set("msg", "Hello 01"))
		suite.Assert().Nil(err, "Failed to write to stream")
		err = stream.Write(logger.NewRecord().Set("bello", "mata banana").Set("level", logger.ERROR).Set("time", time.Now()).Set("msg", "Hello 02"))
		suite.Assert().Nil(err, "Failed to write to stream")
		stream.Flush()
	}
}

func (suite *StreamSuite) TestCanStreamToStackDriverWithKeyFilename() {
	if _, ok := os.LookupEnv("GOOGLE_PROJECT_ID"); !ok {
		suite.T().Skip("There is no way to test this without a Google Project ID")
	}
	if current, ok := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS"); ok {
		_ = os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
		defer func() { _ = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", current) }()
	}
	stream := &logger.StackDriverStream{LogID: "test", KeyFilename: "gcloud-key.json"}
	defer stream.Close()
	suite.Assert().Equal("Stream to Google StackDriver", stream.String())
	suite.Assert().Truef(stream.ShouldWrite(logger.WARN, "", ""), "It should be possible to write to a %s", stream)
	err := stream.Write(logger.NewRecord().Set("bello", "banana").Set("level", logger.WARN).Set("time", time.Now()).Set("msg", "Hello with key filename"))
	suite.Require().Nil(err, "Failed to write to stream")
	stream.Flush()
}

func (suite *StreamSuite) TestCanStreamToStackDriverWithKey() {
	if _, ok := os.LookupEnv("GOOGLE_PROJECT_ID"); !ok {
		suite.T().Skip("There is no way to test this without a Google Project ID")
	}
	if current, ok := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS"); ok {
		_ = os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
		defer func() { _ = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", current) }()
	}
	key := map[string]string{}
	err := Load("../gcloud-key.json", &key)
	suite.Require().Nil(err, "Failed to load Google Cloud key")
	stream := &logger.StackDriverStream{LogID: "test", Key: key}
	defer stream.Close()
	suite.Assert().Equal("Stream to Google StackDriver", stream.String())
	suite.Assert().Truef(stream.ShouldWrite(logger.WARN, "", ""), "It should be possible to write to a %s", stream)
	err = stream.Write(logger.NewRecord().Set("bello", "banana").Set("level", logger.WARN).Set("time", time.Now()).Set("msg", "Hello with key filename"))
	suite.Require().Nil(err, "Failed to write to stream")
	stream.Flush()
}

func (suite *StreamSuite) TestCanStreamToMultiStream() {
	output := CaptureStdout(func() {
		stream := logger.CreateStream(logger.NewLevelSet(logger.INFO), "stdout", "nil")
		suite.Assert().IsType(&logger.MultiStream{}, stream)
		suite.Assert().Equal("MultiStream", fmt.Sprintf("%s", stream))
		suite.Assert().Truef(stream.ShouldWrite(logger.WARN, "", ""), "It should be possible to write to a %s", stream)
		err := stream.Write(logger.NewRecord().Set("bello", "banana").Set("level", logger.ERROR))
		suite.Assert().Nil(err, "Failed to write to stream")
		stream.Flush()
	})
	suite.Assert().JSONEq(`{"bello":"banana","level":50}`, output)
}

func (suite *StreamSuite) TestCanGetFlushFrequencyFromEnvironment() {
	_ = os.Unsetenv("LOG_FLUSHFREQUENCY")
	frequency := logger.GetFlushFrequencyFromEnvironment()
	suite.Assert().Equal(5*time.Minute, frequency, "Frequency should be 5 minutes before being set in the environment")
	_ = os.Setenv("LOG_FLUSHFREQUENCY", "3600s")
	frequency = logger.GetFlushFrequencyFromEnvironment()
	suite.Assert().Equal(1*time.Hour, frequency, "Frequency should be 1 hour after being set in the environment (was %s)", frequency)
	_ = os.Setenv("LOG_FLUSHFREQUENCY", "P2H")
	frequency = logger.GetFlushFrequencyFromEnvironment()
	suite.Assert().Equal(2*time.Hour, frequency, "Frequency should be 2 hour after being set in the environment (was %s)", frequency)
	_ = os.Unsetenv("LOG_FLUSHFREQUENCY")
}

func (suite *StreamSuite) TestFailsWritingToStackDriverWithNoParent() {
	if current, ok := os.LookupEnv("GOOGLE_PROJECT_ID"); ok {
		_ = os.Unsetenv("GOOGLE_PROJECT_ID")
		defer func() { _ = os.Setenv("GOOGLE_PROJECT_ID", current) }()
	}
	stream := &logger.StackDriverStream{}
	err := stream.Write(logger.NewRecord().Set("key", "value"))
	suite.Require().NotNil(err, "Should have failed writing to stream")
	suite.Assert().True(errors.Is(err, errors.EnvironmentMissing), "error should be an Environment Missing error")
	details := errors.EnvironmentMissing.Clone()
	suite.Require().ErrorAs(err, &details, "Error chain should contain an errors.EnvironmentMissing")
	suite.Assert().Equal("GOOGLE_PROJECT_ID", details.What, "Error's What is wrong")
}

func (suite *StreamSuite) TestFailsWritingToStackDriverWithNoCredentials() {
	if _, ok := os.LookupEnv("GOOGLE_PROJECT_ID"); !ok {
		suite.T().Skip("There is no way to test this without a Google Project ID")
	}
	if current, ok := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS"); ok {
		_ = os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
		defer func() { _ = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", current) }()
	}
	stream := &logger.StackDriverStream{}
	err := stream.Write(logger.NewRecord().Set("key", "value"))
	suite.Require().NotNil(err, "Should have failed writing to stream")
	suite.Assert().Contains(err.Error(), "could not find default credentials")
}

func (suite *StreamSuite) TestFailsWritingToStackDriverWithInvalidKey() {
	if _, ok := os.LookupEnv("GOOGLE_PROJECT_ID"); !ok {
		suite.T().Skip("There is no way to test this without a Google Project ID")
	}
	if current, ok := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS"); ok {
		_ = os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
		defer func() { _ = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", current) }()
	}
	stream := &logger.StackDriverStream{LogID: "test", Key: make(chan bool)}
	defer stream.Close()
	err := stream.Write(logger.NewRecord().Set("bello", "banana").Set("level", logger.WARN).Set("time", time.Now()).Set("msg", "Hello with key filename"))
	suite.Require().NotNil(err, "Should have failed writing to stream")
	suite.T().Logf("Expected error: %s", err.Error())
	suite.Assert().Contains(err.Error(), "json: unsupported type")

	stream = &logger.StackDriverStream{LogID: "test", Key: BogusValue{}}
	defer stream.Close()
	err = stream.Write(logger.NewRecord().Set("bello", "banana").Set("level", logger.WARN).Set("time", time.Now()).Set("msg", "Hello with key filename"))
	suite.Require().NotNil(err, "Should have failed writing to stream")
	suite.T().Logf("Expected error: %s", err.Error())
	suite.Assert().Contains(err.Error(), "Not Implemented")
}

func (suite *StreamSuite) TestFailsWritingToFileStreamWithInvalidFile() {
	streamFile := &logger.FileStream{Path: ""}
	err := streamFile.Write(logger.NewRecord().Set("key", "value"))
	suite.Require().NotNil(err, "Should have failed writing to stream")
	suite.Assert().Contains(err.Error(), "no such file")
	var details *os.PathError
	suite.Require().True(errors.As(err, &details), "Error should have been a os.PathError")
	suite.Assert().Equal("open", details.Op)
	suite.Assert().Equal(streamFile.Path, details.Path)
}

func (suite *StreamSuite) TestFailsWritingToFileStreamWithForbiddenPath() {
	streamFile := &logger.FileStream{Path: "/x/test.log"}
	err := streamFile.Write(logger.NewRecord().Set("key", "value"))
	suite.Require().NotNil(err, "Should have failed writing to stream")
	switch runtime.GOOS {
	case "linux", "freebsd":
		suite.Assert().Contains(err.Error(), "permission denied")
	case "darwin":
		suite.Assert().Contains(err.Error(), "read-only file system")
	}
	var details *os.PathError
	suite.Require().True(errors.As(err, &details), "Error should have been a os.PathError")
	suite.Assert().Equal("mkdir", details.Op)
	suite.Assert().Equal(path.Dir(streamFile.Path), details.Path)
}

func (suite *StreamSuite) TestFailsWritingtoMultiStreamWithBogusStream() {
	stream := logger.CreateMultiStream(&logger.StdoutStream{}, &BogusStream{})
	suite.Assert().IsType(&logger.MultiStream{}, stream)
	err := stream.Write(logger.NewRecord().Set("bello", "banana").Set("だれ", "Me"))
	suite.Require().NotNil(err, "Should have failed writing to stream")
	suite.T().Logf("(Expected) Error: %s", err.Error())
}

func (suite *StreamSuite) TestCanFilterMore() {
	streamFile := &logger.FileStream{FilterLevels: logger.NewLevelSet(logger.INFO)}
	streamFile.FilterMore()
	suite.Assert().Equal(logger.WARN, streamFile.FilterLevels.GetDefault())

	streamNil := &logger.NilStream{}
	streamNil.FilterMore()

	streamStackDriver := &logger.StackDriverStream{FilterLevels: logger.NewLevelSet(logger.INFO)}
	streamStackDriver.FilterMore()
	suite.Assert().Equal(logger.WARN, streamStackDriver.FilterLevels.GetDefault())

	streamStderr := &logger.StderrStream{FilterLevels: logger.NewLevelSet(logger.INFO)}
	streamStderr.FilterMore()
	suite.Assert().Equal(logger.WARN, streamStderr.FilterLevels.GetDefault())

	streamStdout := &logger.StdoutStream{FilterLevels: logger.NewLevelSet(logger.INFO)}
	streamStdout.FilterMore()
	suite.Assert().Equal(logger.WARN, streamStdout.FilterLevels.GetDefault())

	streamMulti := logger.CreateMultiStream(streamStderr, streamStdout)
	modifier, ok := streamMulti.(logger.FilterModifier)
	suite.Require().True(ok, "MultiStream should implement FilterModifier")
	modifier.FilterMore()
	suite.Assert().Equal(logger.ERROR, streamStderr.FilterLevels.GetDefault())
	suite.Assert().Equal(logger.ERROR, streamStdout.FilterLevels.GetDefault())
}

func (suite *StreamSuite) TestCanFilterLess() {
	streamFile := &logger.FileStream{FilterLevels: logger.NewLevelSet(logger.INFO)}
	streamFile.FilterLess()
	suite.Assert().Equal(logger.DEBUG, streamFile.FilterLevels.GetDefault())

	streamNil := &logger.NilStream{}
	streamNil.FilterLess()

	streamStackDriver := &logger.StackDriverStream{FilterLevels: logger.NewLevelSet(logger.INFO)}
	streamStackDriver.FilterLess()
	suite.Assert().Equal(logger.DEBUG, streamStackDriver.FilterLevels.GetDefault())

	streamStderr := &logger.StderrStream{FilterLevels: logger.NewLevelSet(logger.INFO)}
	streamStderr.FilterLess()
	suite.Assert().Equal(logger.DEBUG, streamStderr.FilterLevels.GetDefault())

	streamStdout := &logger.StdoutStream{FilterLevels: logger.NewLevelSet(logger.INFO)}
	streamStdout.FilterLess()
	suite.Assert().Equal(logger.DEBUG, streamStdout.FilterLevels.GetDefault())

	streamMulti := logger.CreateMultiStream(streamStderr, streamStdout)
	modifier, ok := streamMulti.(logger.FilterModifier)
	suite.Require().True(ok, "MultiStream should implement FilterModifier")
	modifier.FilterLess()
	suite.Assert().Equal(logger.TRACE, streamStderr.FilterLevels.GetDefault())
	suite.Assert().Equal(logger.TRACE, streamStdout.FilterLevels.GetDefault())
}

func (suite *StreamSuite) TestCanSetLevelPerTopic() {
	streams := []logger.Streamer{
		&logger.FileStream{FilterLevels: logger.NewLevelSet(logger.INFO)},
		&logger.StackDriverStream{FilterLevels: logger.NewLevelSet(logger.INFO)},
		&logger.StderrStream{FilterLevels: logger.NewLevelSet(logger.INFO)},
		&logger.StdoutStream{FilterLevels: logger.NewLevelSet(logger.INFO)},
	}

	for _, strm := range streams {
		setter, ok := strm.(logger.FilterSetter)
		suite.Require().Truef(ok, "Streamer %s should implement FilterSetter", reflect.TypeOf(strm))

		suite.Assert().Truef(strm.ShouldWrite(logger.INFO, "main", ""), "Stream %s should write INFO messages for main topic before it is configured", reflect.TypeOf(strm))
		suite.Assert().Falsef(strm.ShouldWrite(logger.DEBUG, "main", ""), "Stream %s should not write DEBUG messages for main topic before it is configured", reflect.TypeOf(strm))

		setter.SetFilterLevel(logger.DEBUG, "main")

		suite.Assert().Truef(strm.ShouldWrite(logger.WARN, "", ""), "Stream %s should write WARN messages", reflect.TypeOf(strm))
		suite.Assert().Falsef(strm.ShouldWrite(logger.DEBUG, "", ""), "Stream %s should not write DEBUG messages", reflect.TypeOf(strm))

		suite.Assert().Truef(strm.ShouldWrite(logger.DEBUG, "main", ""), "Stream %s should write DEBUG messages for main topic", reflect.TypeOf(strm))
		suite.Assert().Falsef(strm.ShouldWrite(logger.TRACE, "main", ""), "Stream %s should not write TRACE messages for main topic", reflect.TypeOf(strm))

		suite.Assert().Truef(strm.ShouldWrite(logger.INFO, "another_topic", ""), "Stream %s should write INFO messages for another_topic topic", reflect.TypeOf(strm))
		suite.Assert().Falsef(strm.ShouldWrite(logger.DEBUG, "another_topic", ""), "Stream %s should not write DEBUG messages for another_topic topic", reflect.TypeOf(strm))
	}

	streamNil := &logger.NilStream{}
	streamNil.SetFilterLevel(logger.DEBUG, "main")
	suite.Assert().Falsef(streamNil.ShouldWrite(logger.DEBUG, "", ""), "Stream %s should not write DEBUG messages", reflect.TypeOf(streamNil))
	suite.Assert().Falsef(streamNil.ShouldWrite(logger.DEBUG, "main", ""), "Stream %s should not write DEBUG messages for main topic", reflect.TypeOf(streamNil))
	suite.Assert().Falsef(streamNil.ShouldWrite(logger.INFO, "another_topic", ""), "Stream %s should not write INFO messages for another_topic topic", reflect.TypeOf(streamNil))

	streamMulti := logger.CreateMultiStream(&logger.StdoutStream{FilterLevels: logger.NewLevelSet(logger.INFO)}, &logger.FileStream{FilterLevels: logger.NewLevelSet(logger.DEBUG)}).(*logger.MultiStream)
	streamMulti.SetFilterLevel(logger.TRACE, "main")
	suite.Assert().Falsef(streamMulti.ShouldWrite(logger.TRACE, "", ""), "Stream %s should not write TRACE messages", reflect.TypeOf(streamMulti))
	suite.Assert().Truef(streamMulti.ShouldWrite(logger.DEBUG, "", ""), "Stream %s should write DEBUG messages", reflect.TypeOf(streamMulti))
	suite.Assert().Truef(streamMulti.ShouldWrite(logger.DEBUG, "main", ""), "Stream %s should write DEBUG messages for main topic", reflect.TypeOf(streamMulti))
	suite.Assert().Truef(streamMulti.ShouldWrite(logger.INFO, "another_topic", ""), "Stream %s should write INFO messages for another_topic topic", reflect.TypeOf(streamMulti))
}

func (suite *StreamSuite) TestCanSetLevelPerTopicAndScope() {
	streams := []logger.Streamer{
		&logger.FileStream{FilterLevels: logger.NewLevelSet(logger.INFO)},
		&logger.StackDriverStream{FilterLevels: logger.NewLevelSet(logger.INFO)},
		&logger.StderrStream{FilterLevels: logger.NewLevelSet(logger.INFO)},
		&logger.StdoutStream{FilterLevels: logger.NewLevelSet(logger.INFO)},
	}

	for _, strm := range streams {
		setter, ok := strm.(logger.FilterSetter)
		suite.Require().Truef(ok, "Streamer %s should implement FilterSetter", reflect.TypeOf(strm))

		suite.Assert().Truef(strm.ShouldWrite(logger.INFO, "main", "any"), "Stream %s should write INFO messages for main topic and any scope before it is configured", reflect.TypeOf(strm))
		suite.Assert().Falsef(strm.ShouldWrite(logger.DEBUG, "main", "any"), "Stream %s should not write DEBUG messages for main topic and any scope before it is configured", reflect.TypeOf(strm))

		setter.SetFilterLevel(logger.TRACE, "main", "specific")
		setter.SetFilterLevel(logger.DEBUG, "main")

		suite.Assert().Truef(strm.ShouldWrite(logger.DEBUG, "main", "any"), "Stream %s should write DEBUG messages for main topic and any scope", reflect.TypeOf(strm))
		suite.Assert().Truef(strm.ShouldWrite(logger.TRACE, "main", "specific"), "Stream %s should write TRACE messages for main topic and specific scope", reflect.TypeOf(strm))
		suite.Assert().Falsef(strm.ShouldWrite(logger.TRACE, "main", "any"), "Stream %s should not write TRACE messages for main topic and any scope", reflect.TypeOf(strm))

		suite.Assert().Falsef(strm.ShouldWrite(logger.DEBUG, "another_topic", "any"), "Stream %s should not write DEBUG messages for another_topic topic and any scope", reflect.TypeOf(strm))
	}

	streamNil := &logger.NilStream{}
	streamNil.SetFilterLevel(logger.TRACE, "main", "specific")
	suite.Assert().Falsef(streamNil.ShouldWrite(logger.DEBUG, "main", "any"), "Stream %s should not write DEBUG messages for main topic and any scope", reflect.TypeOf(streamNil))
	suite.Assert().Falsef(streamNil.ShouldWrite(logger.TRACE, "main", "specific"), "Stream %s should not write TRACE messages for main topic and specific scope", reflect.TypeOf(streamNil))
	suite.Assert().Falsef(streamNil.ShouldWrite(logger.DEBUG, "another_topic", "any"), "Stream %s should not write DEBUG messages for another_topic topic and any scope", reflect.TypeOf(streamNil))

	streamMulti := logger.CreateMultiStream(&logger.StdoutStream{FilterLevels: logger.NewLevelSet(logger.INFO)}, &logger.FileStream{FilterLevels: logger.NewLevelSet(logger.DEBUG)}).(*logger.MultiStream)
	streamMulti.SetFilterLevel(logger.TRACE, "main", "specific")
	suite.Assert().Falsef(streamMulti.ShouldWrite(logger.TRACE, "main", "any"), "Stream %s should not write TRACE messages for main topic and any scope", reflect.TypeOf(streamMulti))
	suite.Assert().Truef(streamMulti.ShouldWrite(logger.DEBUG, "main", "any"), "Stream %s should write DEBUG messages for main topic and any scope", reflect.TypeOf(streamMulti))
	suite.Assert().Truef(streamMulti.ShouldWrite(logger.TRACE, "main", "specific"), "Stream %s should write TRACE messages for main topic and specific scope", reflect.TypeOf(streamMulti))
	suite.Assert().Truef(streamMulti.ShouldWrite(logger.DEBUG, "another_topic", "any"), "Stream %s should write DEBUG messages for another_topic topic and any scope", reflect.TypeOf(streamMulti))
}

func (suite *StreamSuite) TestCanCreateWithEnvironmentPrefix() {
	_ = os.Setenv("LOG_DESTINATION", "stdout")
	defer func() { _ = os.Unsetenv("LOG_DESTINATION") }()
	_ = os.Setenv("TEST_LOG_DESTINATION", "nil")
	defer func() { _ = os.Unsetenv("TEST_LOG_DESTINATION") }()
	stream := logger.CreateStreamWithPrefix(logger.EnvironmentPrefix("TEST_"), logger.NewLevelSet(logger.INFO))
	suite.Require().NotNil(stream, "Failed to create a stream from an empty destination")
	suite.Assert().IsType(&logger.NilStream{}, stream)
	stream.Close()
}
