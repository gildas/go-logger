package logger_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"testing"

	"github.com/gildas/go-logger"
	"github.com/stretchr/testify/require"
)

type BogusStream struct {
}

// SetFilterLevel sets the filter level
func (stream *BogusStream) SetFilterLevel(level logger.Level) {
}

// SetFilterLevelIfUnset sets the filter level if not set already
func (stream *BogusStream) SetFilterLevelIfUnset(level logger.Level) {
}

func (stream *BogusStream) Write(record logger.Record) error {
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

type BogusValue struct {
}

func (v *BogusValue) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("Failed to Marshal BogusValue")
}

// Load loads an object from a file and marshals it
func Load(filename string, object interface{}) (err error) {
	var payload []byte

	if payload, err = ioutil.ReadFile(filepath.Join(".", "testdata", filename)); err != nil {
		return
	}
	if err = json.Unmarshal(payload, object); err != nil {
		return
	}
	return
}

// RequireEqualJSON tests if an unmarshaled object matches the JSON stored in the given file
func RequireEqualJSON(t *testing.T, filename string, payload []byte) {
	expected, err := ioutil.ReadFile(filepath.Join(".", "testdata", filename))
	require.Nil(t, err, "Failed to load %s", filename)
	require.JSONEq(t, string(expected), string(payload))
}

// CreateLogger creates a new logger in a temp destination
func CreateLogger(t *testing.T, filename string, wantLocal bool) (*logger.Logger, func()) {
	var folder string
	var teardown func()

	if wantLocal {
		folder, teardown = CreateLogDir(t)
	} else {
		folder, teardown = CreateTempDir(t)
	}
	path := filepath.Join(folder, filename)
	log := logger.CreateWithDestination("test", "file://"+path)
	//if _, err := os.Stat(path); err != nil {
	//	t.Fatalf("Log file was not created at: %s. Error: %s\n", path, err)
	//}
	return log, func() { log.Close(); teardown() }
}

// CreateTempDir creates a temporary directory
// return the temp folder and a func to delete it when done
func CreateTempDir(t *testing.T) (string, func()) {
	dir, err := ioutil.TempDir("", "go_logger")
	if err != nil {
		t.Fatalf("Unable to create a temp folder for log files. Error: %s\n", err)
	}
	return dir, func() { os.RemoveAll(dir) }
}

// CreateLogDir creates a local log directory
func CreateLogDir(t *testing.T) (string, func()) {
	dir := filepath.Join(".", "log")
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		t.Fatalf("Unable to create log folder for log files. Error: %s\n", err)
	}
	return dir, func() {}
}

func CaptureStderr(f func()) string {
	reader, writer, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	stderr := os.Stderr
	os.Stderr = writer
	defer func() {
		os.Stderr = stderr
	}()

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
					stringvalue  = fmt.Sprintf("%v", value)
				default:
					suite.Failf(fmt.Sprintf("The value of the key %s cannot be casted to string", key), "Type: %s", reflect.TypeOf(value))
			}
			suite.Assert().Truef(rex.MatchString(stringvalue), "Key %s: the value %v does not match the regex %s", key, value, rex)
		}
	}

	for key := range properties {
		suite.Assert().Contains(rex_records, key, "The line contains the extra key %s", key)
	}
}
