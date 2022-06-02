package logger

import "time"

// TimeFunc will log the duration of the given function call
func (l *Logger) TimeFunc(message string, code func()) {
	start := time.Now()
	code()
	duration := time.Since(start)
	l.Record("duration", duration.Seconds()).Infof("%s. executed in %s", message, duration)
}

// TimeFuncV will log the duration of the given function call that returns a value
func (l *Logger) TimeFuncV(message string, code func() interface{}) interface{} {
	start := time.Now()
	result := code()
	duration := time.Since(start)
	l.Record("duration", duration.Seconds()).Infof("%s. executed in %s", message, duration)
	return result
}

// TimeFuncE will log the duration of the given function call that returns an error
func (l *Logger) TimeFuncE(message string, code func() error) error {
	start := time.Now()
	err := code()
	duration := time.Since(start)
	l.Record("duration", duration.Seconds()).Infof("%s. executed in %s", message, duration)
	return err
}

// TimeFuncVE will log the duration of the given function call that returns a value and an error
func (l *Logger) TimeFuncVE(message string, code func() (interface{}, error)) (interface{}, error) {
	start := time.Now()
	result, err := code()
	duration := time.Since(start)
	l.Record("duration", duration.Seconds()).Infof("%s. executed in %s", message, duration)
	return result, err
}
