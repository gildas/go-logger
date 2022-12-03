package logger

import "regexp"

type Redactor regexp.Regexp

func NewRedactor(regex string) (*Redactor, error) {
	r, err := regexp.Compile(regex)
	return (*Redactor)(r), err
}

func (redactor Redactor) Redact(value string) (string, bool) {
	// TODO: Find a way to not run the regex twice
	if (*regexp.Regexp)(&redactor).MatchString(value) {
		return (*regexp.Regexp)(&redactor).ReplaceAllString(value, "REDACTED"), true
	}
	return value, false
}

// String return a String representation of a Message
//
// implements fmt.Stringer
func (redactor Redactor) String() string {
	return (*regexp.Regexp)(&redactor).String()
}
