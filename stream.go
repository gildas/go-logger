package logger

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Streamer is the interface a Logger writes to
type Streamer interface {
	Write(record Record) error
	ShouldWrite(level Level) bool
	Flush()
}

// GetFlushFrequencyFromEnvironment fetches the flush frequency from the environment
//  the frequency should be like https://golang.org/pkg/time/#ParseDuration or an ISO8601 duration.  
//
//  If not set, the frequency will be 5 minutes
func GetFlushFrequencyFromEnvironment() time.Duration {
	if value, ok := os.LookupEnv("LOG_FLUSHFREQUENCY"); ok {
		if strings.HasPrefix(value, "P") {
			if duration, err := parseDuration(value); err == nil {
				return duration
			}
		} else {
			if !strings.HasSuffix(value, "h") && !strings.HasSuffix(value, "m") && !strings.HasSuffix(value, "s") {
				value = value + "s"
			}
			if duration, err := time.ParseDuration(value); err == nil {
				return duration
			}
		}
	}
	return 5 * time.Minute
}

// CreateStreamWithDestination creates a new Streamer from a list of strings
func CreateStreamWithDestination(destinations ...string) Streamer {
	if len(destinations) == 0 {
		destination, ok := os.LookupEnv("LOG_DESTINATION")
		if !ok || len(destination) == 0 {
			return &StdoutStream{}
		}
		destinations = []string{destination}
	}
	streams := []Streamer{}

	for _, destination := range destinations {
		var stream Streamer
		switch strings.ToLower(destination) {
		case "stdout":
			stream = &StdoutStream{}
		case "stderr":
			stream = &StderrStream{}
		case "gcp", "google", "googlecloud":
			stream = &GCPStream{}
		case "stackdriver":
			stream =  &StackDriverStream{}
		case "nil", "null", "void", "blackhole", "nether":
			stream = &NilStream{}
		default:
			if strings.HasPrefix(destination, "file://") {
				stream = &FileStream{Path: strings.TrimPrefix(destination, "file://")}
			} else if len(destination) > 0 {
				stream = &FileStream{Path: destination}
			} else {
				stream = &StdoutStream{}
			}
		}
		streams = append(streams, stream)
	}
	if len(streams) == 1 {
		return streams[0]
	}
	return &MultiStream{ streams: streams }
}

// We cannot import github.com/gildas/go-core... so we have to duplicate the function here...

// parseDuration parses an ISO8601 duration
func parseDuration(iso8601 string) (duration time.Duration, err error) {
	parser := regexp.MustCompile(`P(?P<years>\d+Y)?(?P<months>\d+M)?(?P<weeks>\d+W)?(?P<days>\d+D)?T?(?P<hours>\d+H)?(?P<minutes>\d+M)?(?P<seconds>\d+S)?`)
	matches := parser.FindStringSubmatch(iso8601)
	var parsed int

	if len(matches) == 0 {
		return time.Duration(0), errors.Errorf(`"%s" is not an ISO8601 duration`, iso8601)
	}
	if len(matches[1]) > 0 {
		if parsed, err = strconv.Atoi(matches[1][:len(matches[1])-1]); err != nil {
			return
		}
		duration = time.Duration(parsed*24*365) * time.Hour // years
	}

	if len(matches[2]) > 0 {
		if parsed, err = strconv.Atoi(matches[2][:len(matches[2])-1]); err != nil {
			return
		}
		duration += time.Duration(parsed*24*30) * time.Hour // months
	}

	if len(matches[3]) > 0 {
		if parsed, err = strconv.Atoi(matches[3][:len(matches[3])-1]); err != nil {
			return
		}
		duration += time.Duration(parsed*24*7) * time.Hour // weeks
	}

	if len(matches[4]) > 0 {
		if parsed, err = strconv.Atoi(matches[4][:len(matches[4])-1]); err != nil {
			return
		}
		duration += time.Duration(parsed*24) * time.Hour // days
	}

	if len(matches[5]) > 0 {
		if parsed, err = strconv.Atoi(matches[5][:len(matches[5])-1]); err != nil {
			return
		}
		duration += time.Duration(parsed) * time.Hour // hours
	}

	if len(matches[6]) > 0 {
		if parsed, err = strconv.Atoi(matches[6][:len(matches[6])-1]); err != nil {
			return
		}
		duration += time.Duration(parsed) * time.Minute // minutes
	}

	if len(matches[7]) > 0 {
		if parsed, err = strconv.Atoi(matches[7][:len(matches[7])-1]); err != nil {
			return
		}
		duration += time.Duration(parsed) * time.Second // seconds
	}

	return
}
