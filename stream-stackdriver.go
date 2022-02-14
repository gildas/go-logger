package logger

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/logging"
	"github.com/gildas/go-errors"
	googleoption "google.golang.org/api/option"
)

// StackDriverStream is the Stream that writes to the standard output
//
// implements logger.Streamer
type StackDriverStream struct {
	LogID       string
	Parent      string
	KeyFilename string
	Key         interface{}
	Converter   Converter
	FilterLevel Level
	mutex       sync.Mutex
	client      *logging.Client
	target      *logging.Logger
}

// SetFilterLevel sets the filter level
//
// implements logger.FilterSetter
func (stream *StackDriverStream) SetFilterLevel(level Level) {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	stream.FilterLevel = level
}

// SetFilterLevelIfUnset sets the filter level if not set already
//
// implements logger.FilterSetter
func (stream *StackDriverStream) SetFilterLevelIfUnset(level Level) {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	if stream.FilterLevel == UNSET {
		stream.FilterLevel = level
	}
}

// Write writes the given Record
//
// implements logger.Streamer
func (stream *StackDriverStream) Write(record Record) (err error) {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	if stream.client == nil {
		ctx := context.Background()
		if len(stream.Parent) == 0 {
			projectID, ok := os.LookupEnv("GOOGLE_PROJECT_ID")
			if !ok {
				return errors.EnvironmentMissing.With("GOOGLE_PROJECT_ID")
			}
			stream.Parent = "projects/" + projectID
		}
		stream.Converter = &StackDriverConverter{}
		options := []googleoption.ClientOption{}
		if stream.Key != nil {
			payload, err := json.Marshal(stream.Key)
			if err != nil {
				return errors.JSONMarshalError.Wrap(err)
			}
			options = append(options, googleoption.WithCredentialsJSON(payload))
		} else if len(stream.KeyFilename) != 0 {
			options = append(options, googleoption.WithCredentialsFile(stream.KeyFilename))
		}
		// if Key and KeyFilename are not provided, $GOOGLE_APPLICATION_CREDENTIALS is used.
		stream.client, err = logging.NewClient(ctx, stream.Parent, options...)
		if err != nil {
			return errors.WithStack(err)
		}
		stream.target = stream.client.Logger(stream.LogID)
		if stream.FilterLevel == 0 {
			stream.FilterLevel = GetLevelFromEnvironment()
		}
	}
	grecord := stream.Converter.Convert(record)
	stamp, _ := time.Parse(time.RFC3339, grecord["time"].(string))
	severity := grecord["severity"].(logging.Severity)
	delete(grecord, "time")
	delete(grecord, "severity")
	stream.target.Log(logging.Entry{
		Timestamp: stamp,
		Severity:  severity,
		Payload:   grecord,
	})
	return nil
}

// ShouldWrite tells if the given level should be written to this stream
//
// implements logger.Streamer
func (stream *StackDriverStream) ShouldWrite(level Level) bool {
	return level.ShouldWrite(stream.FilterLevel)
}

// Flush flushes the stream (makes sure records are actually written)
//
// implements logger.Streamer
func (stream *StackDriverStream) Flush() {
	if stream.target != nil {
		stream.mutex.Lock()
		defer stream.mutex.Unlock()
		stream.target.Flush()
	}
}

// Close closes the stream
//
// implements logger.Streamer
func (stream *StackDriverStream) Close() {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	if stream.target != nil {
		stream.target.Flush()
	}
	if stream.client != nil {
		stream.client.Close()
	}
}

// String gets a string version
//
// implements fmt.Stringer
func (stream *StackDriverStream) String() string {
	return "Stream to Google StackDriver"
}
