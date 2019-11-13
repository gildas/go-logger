package logger

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pkg/errors"
)

// FileStream is the Stream that writes to a file
//   Any record with a level < FilterLevel will be written
type FileStream struct {
	*json.Encoder
	FilterLevel Level
	Path        string
	File        *os.File
}

// Write writes the given Record
//   implements logger.Stream
func (stream *FileStream) Write(record Record) (err error) {
	if stream.File == nil {
		const flags = os.O_CREATE | os.O_APPEND | os.O_WRONLY
		const perms = 0644
		if stream.File, err = os.OpenFile(stream.Path, flags, perms); err != nil {
			return errors.WithStack(err)
		}
		if stream.Encoder == nil {
			stream.Encoder = json.NewEncoder(stream.File)
		}
		if stream.FilterLevel == 0 {
			stream.FilterLevel = GetLevelFromEnvironment()
		}
	}
	if err := stream.Encoder.Encode(record); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// ShouldWrite tells if the given level should be written to this stream
//   implements logger.Stream
func (stream *FileStream) ShouldWrite(level Level) bool {
	return level.ShouldWrite(stream.FilterLevel)
}

// Flush flushes the stream (makes sure records are actually written)
//   implements logger.Stream
func (stream *FileStream) Flush() {
}

// String gets a string version
//   implements the fmt.Stringer interface
func (stream FileStream) String() string {
	return fmt.Sprintf("Stream to %s", stream.Path)
}