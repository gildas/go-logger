package logger

import (
	"regexp"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

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