package logger

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/pkg/errors"
)

// FileStream is the Stream that writes to a file
//   Any record with a level < FilterLevel will be written
type FileStream struct {
	*json.Encoder
	Path        string
	FilterLevel Level
	Unbuffered  bool
	file        *os.File
	output      *bufio.Writer
}

// Write writes the given Record
//   implements logger.Stream
func (stream *FileStream) Write(record Record) (err error) {
	if stream.file == nil {
		const flags = os.O_CREATE | os.O_APPEND | os.O_WRONLY
		const perms = 0644
		if stream.file, err = os.OpenFile(stream.Path, flags, perms); err != nil {
			return errors.WithStack(err)
		}
		if stream.Unbuffered {
			stream.output =  nil
			stream.Encoder = json.NewEncoder(stream.file)
		} else {
			stream.output = bufio.NewWriter(stream.file)
			stream.Encoder = json.NewEncoder(stream.output)
		}
		if stream.FilterLevel == 0 {
			stream.FilterLevel = GetLevelFromEnvironment()
		}
	}
	if err := stream.Encoder.Encode(record); err != nil {
		return errors.WithStack(err)
	}
	if GetLevelFromRecord(record) >= ERROR {
		stream.Flush()
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
	if stream.output != nil {
		stream.output.Flush()
	}
}

// String gets a string version
//   implements the fmt.Stringer interface
func (stream FileStream) String() string {
	if stream.Unbuffered {
		return fmt.Sprintf("Unbuffered Stream to %s, Filter: %s", stream.Path, stream.FilterLevel)
	}
	return fmt.Sprintf("Stream to %s, Filter: %s", stream.Path, stream.FilterLevel)
}