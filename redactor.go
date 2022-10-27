package logger

import "regexp"

type Redactor regexp.Regexp

func NewRedactor(regex string) (*Redactor, error) {
	r, err := regexp.Compile(regex)
	return (*Redactor)(r), err
}

func (redactor Redactor) Redact(value string) (string, bool) {
	redacted := ((*regexp.Regexp)(&redactor)).ReplaceAllString(value, "REDACTED")
	return redacted, true
}

// String return a String representation of a Message
//
// implements fmt.Stringer
func (redactor Redactor) String() string {
	return (*regexp.Regexp)(&redactor).String()
}
