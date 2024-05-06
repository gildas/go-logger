package logger_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"testing"

	"github.com/gildas/go-errors"
	"github.com/gildas/go-logger"
	"github.com/stretchr/testify/require"
)

type BogusStream struct {
}

// GetFilterLevels gets the filter level
func (stream *BogusStream) GetFilterLevels() logger.LevelSet {
	return logger.LevelSet{}
}

// SetFilterLevel sets the filter level
func (stream *BogusStream) SetFilterLevel(level logger.Level) {
}

// SetFilterLevelIfUnset sets the filter level if not set already
func (stream *BogusStream) SetFilterLevelIfUnset(level logger.Level) {
}

func (stream *BogusStream) Write(record *logger.Record) error {
	return fmt.Errorf("This Stream is Bogus")
}

func (stream *BogusStream) ShouldLogSourceInfo() bool {
	return false
}

func (stream *BogusStream) ShouldWrite(level logger.Level, topic, scope string) bool {
	return true
}

func (stream *BogusStream) Flush() {
}

func (stream *BogusStream) Close() {
}

// BogusValue is a bogus value that fails to marshal
type BogusValue struct {
}

func (v BogusValue) MarshalJSON() ([]byte, error) {
	return nil, errors.Join(errors.JSONMarshalError, errors.NotImplemented)
}

// NonMarshableError is an error that fails to marshal
type NonMarshableError struct {
	Message string
	Channel chan int
}

func (err NonMarshableError) Error() string {
	return err.Message
}

// NonMarshableObject is an object that fails to marshal but implements fmt.Stringer
type NonMarshableObject struct {
	Message string
	Channel chan int
}

func (object NonMarshableObject) String() string {
	return object.Message
}

type User struct {
	ID     string         `json:"id"`
	Name   string         `json:"name"`
	logger *logger.Logger `json:"-"`
}

func (user User) Redact() interface{} {
	return User{user.ID, logger.Redact(user.Name), user.logger}
}

func (user User) GetLogger() *logger.Logger {
	return user.logger
}

func (user User) String() string {
	return user.Name
}

type Metadata struct {
	UserID string `json:"userId"`
	Name string `json:"name"`
	City string `json:"city"`
}

func (metadata Metadata) Redact(keyToRedact ...string) interface{} {
	redact := metadata
	for _, key := range keyToRedact {
		switch key {
		case "userId":
			redact.UserID = logger.Redact(metadata.UserID)
		case "name":
			redact.Name = logger.Redact(metadata.Name)
		case "city":
			redact.City = logger.Redact(metadata.City)
		}
	}
	return &redact
}

// Load loads an object from a file and marshals it
func Load(filename string, object interface{}) (err error) {
	var payload []byte

	if payload, err = os.ReadFile(filepath.Join(".", "testdata", filename)); err != nil {
		return
	}
	if err = json.Unmarshal(payload, object); err != nil {
		return
	}
	return
}

// RequireEqualJSON tests if an unmarshaled object matches the JSON stored in the given file
func RequireEqualJSON(t *testing.T, filename string, payload []byte) {
	expected, err := os.ReadFile(filepath.Join(".", "testdata", filename))
	require.Nil(t, err, "Failed to load %s", filename)
	require.JSONEq(t, string(expected), string(payload))
}

// CreateLogger creates a new logger in a temp destination
func CreateLogger(filename string, wantLocal bool) (*logger.Logger, func()) {
	var folder string
	var teardown func()

	if wantLocal {
		folder, teardown = CreateLogDir()
	} else {
		folder, teardown = CreateTempDir()
	}
	path := filepath.Join(folder, filename)
	log := logger.Create("test", "file://"+path)
	//if _, err := os.Stat(path); err != nil {
	//	t.Fatalf("Log file was not created at: %s. Error: %s\n", path, err)
	//}
	return log, func() { log.Close(); teardown() }
}

// CreateTempDir creates a temporary directory
//
// return the temp folder and a func to delete it when done
func CreateTempDir() (string, func()) {
	dir, err := os.MkdirTemp("", "go_logger")
	if err != nil {
		panic(fmt.Sprintf("Unable to create a temp folder for log files. Error: %s\n", err))
	}
	return dir, func() { os.RemoveAll(dir) }
}

// CreateLogDir creates a local log directory
func CreateLogDir() (string, func()) {
	dir := filepath.Join(".", "log")
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		panic(fmt.Sprintf("Unable to create log folder for log files. Error: %s\n", err))
	}
	return dir, func() {}
}

// CreateTempFile creates a temporary file
//
// return the temp file and a func to delete it when done
func CreateTempFile() (*os.File, func()) {
	file, err := os.CreateTemp("", "go_logger")
	if err != nil {
		panic(fmt.Sprintf("Unable to create a temp file for log files. Error: %s\n", err))
	}
	return file, func() { file.Close(); os.Remove(file.Name()) }
}

func CaptureStderr(f func()) string {
	reader, writer, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	defer func(stderr *os.File) {
		os.Stderr = stderr
	}(os.Stderr)
	os.Stderr = writer

	f()
	writer.Close()

	output := bytes.Buffer{}
	_, _ = io.Copy(&output, reader)
	return output.String()
}

func CaptureStdout(f func()) string {
	reader, writer, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	defer func(stdout *os.File) {
		os.Stdout = stdout
	}(os.Stdout)
	os.Stdout = writer

	f()
	writer.Close()

	output := bytes.Buffer{}
	_, _ = io.Copy(&output, reader)
	return output.String()
}

func (suite *LoggerSuite) LogLineEqual(line string, records map[string]string) {
	rex_records := make(map[string]*regexp.Regexp)
	for key, value := range records {
		rex_records[key] = regexp.MustCompile(value)
	}

	properties := map[string]interface{}{}
	err := json.Unmarshal([]byte(line), &properties)
	suite.Require().NoError(err, "Could not unmarshal line, error: %s", err)

	for key, rex := range rex_records {
		suite.Assert().Contains(properties, key, "The line does not contain the key %s", key)
		if value, found := properties[key]; found {
			var stringvalue string
			switch actual := value.(type) {
			case string:
				stringvalue = actual
			case int, int8, int16, int32, int64:
				stringvalue = strconv.FormatInt(value.(int64), 10)
			case uint, uint8, uint16, uint32, uint64:
				stringvalue = strconv.FormatUint(value.(uint64), 10)
			case float32, float64:
				stringvalue = strconv.FormatFloat(value.(float64), 'f', -1, 64)
			case fmt.Stringer:
				stringvalue = actual.String()
			case map[string]interface{}:
				stringvalue = fmt.Sprintf("%v", value)
			default:
				suite.Failf(fmt.Sprintf("The value of the key %s cannot be casted to string", key), "Type: %s", reflect.TypeOf(value))
			}
			suite.Assert().Truef(rex.MatchString(stringvalue), `Key "%s": the value %v does not match the regex /%s/`, key, value, rex)
		}
	}

	for key := range properties {
		suite.Assert().Contains(rex_records, key, "The line contains the extra key %s", key)
	}
}
