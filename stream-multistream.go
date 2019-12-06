package logger

// MultiStream is the Stream that writes to several streams
type MultiStream struct {
	streams []Streamer
}

// SetFilterLevel sets the filter level of all streams
func (stream *MultiStream) SetFilterLevel(level Level) Streamer {
	for _, s := range stream.streams {
		s.SetFilterLevel(level)
	}
	return stream
}

// Write writes the given Record
//   implements logger.Stream
func (stream *MultiStream) Write(record Record) error {
	for _, s := range stream.streams {
		if err := s.Write(record); err != nil {
			return err
		}
	}
	return nil
}

// ShouldWrite tells if the given level should be written to this stream
//   implements logger.Stream
func (stream *MultiStream) ShouldWrite(level Level) bool {
	return true
}

// Flush flushes the stream (makes sure records are actually written)
//   implements logger.Stream
func (stream *MultiStream) Flush() {
	for _, s := range stream.streams {
		s.Flush()
	}
}

// String gets a string version
//   implements the fmt.Stringer interface
func (stream MultiStream) String() string {
	return "MultiStream"
}

func CreateMultiStream(streams ...Streamer) Streamer {
	return &MultiStream{streams: streams}
}