package logger_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/gildas/go-logger"
	"github.com/stretchr/testify/require"
)

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
	log := logger.CreateWithDestination("test", "file://" + path)
	//if _, err := os.Stat(path); err != nil {
	//	t.Fatalf("Log file was not created at: %s. Error: %s\n", path, err)
	//}
	return log, teardown
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