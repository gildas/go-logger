package logger

import (
	"context"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/logging"
	"github.com/pkg/errors"
)


// GCPStream is the Stream that writes to the standard output
type StackDriverStream struct {
	LogID       string
	ProjectID   string
	FilterLevel Level
	mutex       sync.Mutex
	client      *logging.Client
	target      *logging.Logger
}

// SetFilterLevel sets the filter level
func (stream *StackDriverStream) SetFilterLevel(level Level) Streamer {
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	stream.FilterLevel = level
	return stream
}

// Write writes the given Record
//   implements logger.Stream
func (stream *StackDriverStream) Write(record Record) (err error) {
	if stream.client == nil {
		ctx := context.Background()
		if len(stream.ProjectID) == 0 {
			projectID, ok := os.LookupEnv("PROJECT_ID")
			if !ok {
				return errors.New("Missing environment variable PROJECT_ID")
			}
			stream.ProjectID = projectID
		}
		stream.client, err = logging.NewClient(ctx, stream.ProjectID)
		if err != nil {
			return errors.WithStack(err)
		}
		//defer client.Close()
		stream.target = stream.client.Logger(stream.LogID)
		if stream.FilterLevel == 0 {
			stream.FilterLevel = GetLevelFromEnvironment()
		}
	}
	var stamp time.Time

	if stamp, err = time.Parse(time.RFC3339, record["time"].(string)); err != nil {
		stamp = time.Now()
	}

	delete(record, "level")
	delete(record, "time")
	delete(record, "name")
	stream.target.Log(logging.Entry{
		Timestamp: stamp,
		Severity:  severity(record["level"]),
		Payload:   record,
	})
	return nil
}

// ShouldWrite tells if the given level should be written to this stream
//   implements logger.Stream
func (stream *StackDriverStream) ShouldWrite(level Level) bool {
	return level.ShouldWrite(stream.FilterLevel)
}

// Flush flushes the stream (makes sure records are actually written)
//   implements logger.Stream
func (stream *StackDriverStream) Flush() {
}

// String gets a string version
//   implements the fmt.Stringer interface
func (stream StackDriverStream) String() string {
	return "Stream to Google StackDriver"
}

func severity(level interface{}) logging.Severity {
	switch level.(Level) {
	case NEVER:
		return logging.Default
	case TRACE:
		return logging.Debug
	case DEBUG:
		return logging.Debug
	case INFO:
		return logging.Info
	case WARN:
		return logging.Warning
	case ERROR:
		return logging.Error
	case FATAL:
		return logging.Critical
	case ALWAYS:
		return logging.Emergency
	default:
		return logging.Info
	}
}