package logger

import (
	"context"
	"encoding/json"
	"fmt"
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
	LogID        string
	Parent       string
	KeyFilename  string
	Key          interface{}
	Converter    Converter
	FilterLevels LevelSet
	SourceInfo   bool
	mutex        sync.Mutex
	client       *logging.Client
	target       *logging.Logger
}

// GetFilterLevels gets the filter levels
//
// implements logger.Streamer
func (stream *StackDriverStream) GetFilterLevels() LevelSet {
	return stream.FilterLevels
}

// SetFilterLevel sets the filter level
//
// If present, the first parameter is the topic.
//
// If present, the second parameter is the scope.
//
// implements logger.FilterSetter
func (stream *StackDriverStream) SetFilterLevel(level Level, parameters ...string) {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	if len(parameters) == 0 {
		stream.FilterLevels.SetDefault(level)
	} else if len(parameters) == 1 {
		stream.FilterLevels.Set(level, parameters[0], "")
	} else {
		stream.FilterLevels.Set(level, parameters[0], parameters[1])
	}
}

// FilterMore tells the stream to filter more
//
// The stream will filter more if it is not already at the highest level.
// Which means less log messages will be written to the stream
//
// Example: if the stream is at DEBUG, it will be filtering at INFO
//
// implements logger.FilterModifier
func (stream *StackDriverStream) FilterMore() {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	stream.FilterLevels.SetDefault(stream.FilterLevels.GetDefault().Next())
}

// FilterLess tells the stream to filter less
//
// The stream will filter less if it is not already at the lowest level.
// Which means more log messages will be written to the stream
//
// Example: if the stream is at INFO, it will be filtering at DEBUG
//
// implements logger.FilterModifier
func (stream *StackDriverStream) FilterLess() {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	stream.FilterLevels.SetDefault(stream.FilterLevels.GetDefault().Previous())
}

// Write writes the given Record
//
// implements logger.Streamer
func (stream *StackDriverStream) Write(record *Record) (err error) {
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
			if errors.Is(err, errors.JSONMarshalError) {
				return err
			} else if err != nil {
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
		if len(stream.FilterLevels) == 0 {
			stream.FilterLevels = ParseLevelsFromEnvironment()
		}
	}
	grecord := stream.Converter.Convert(record)
	stamp, _ := time.Parse(time.RFC3339, grecord.Get("time").(string))
	severity := grecord.Get("severity").(logging.Severity)
	grecord.Delete("time")
	grecord.Delete("severity")
	stream.target.Log(logging.Entry{
		Timestamp: stamp,
		Severity:  severity,
		Payload:   grecord,
	})
	return nil
}

// ShouldLogSourceInfo tells if the source info should be logged
//
// implements logger.Streamer
func (stream *StackDriverStream) ShouldLogSourceInfo() bool {
	return stream.SourceInfo
}

// ShouldWrite tells if the given level should be written to this stream
//
// implements logger.Streamer
func (stream *StackDriverStream) ShouldWrite(level Level, topic, scope string) bool {
	return level.ShouldWrite(stream.FilterLevels.Get(topic, scope))
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
	if len(stream.FilterLevels) > 0 {
		return fmt.Sprintf("Stream to Google StackDriver, Filter: %s", stream.FilterLevels)
	}
	return "Stream to Google StackDriver"
}
