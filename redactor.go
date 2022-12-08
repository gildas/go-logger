package logger

import (
	"regexp"

	"github.com/gildas/go-core"
)

type Redactor regexp.Regexp

var (
	// CreditCardRedactor is a Redactor that will redact Credit Card Numbers
	CreditCardRedactor = core.Must(NewRedactor(`(4[0-9]{12}(?:[0-9]{3})?)|((?:5[1-5][0-9]{2}|222[1-9]|22[3-9][0-9]|2[3-6][0-9]{2}|27[01][0-9]|2720)[0-9]{12})|(3[47][0-9]{13})|(3(?:0[0-5]|[68][0-9])[0-9]{11})|(6(?:011|5[0-9]{2})[0-9]{12})|((?:2131|1800|35\d{3})\d{11})`))

	// AMEXRedactor is a Redactor that will redact AMEX Numbers
	AMEXRedactor = core.Must(NewRedactor(`3[47][0-9]{13}`))

	// DinersClubRedactor is a Redactor that will redact Diners Club Numbers
	DinersClubRedactor = core.Must(NewRedactor(`3(?:0[0-5]|[68][0-9])[0-9]{11}`))

	// DiscoverRedactor is a Redactor that will redact Discover Numbers
	DiscoverRedactor = core.Must(NewRedactor(`6(?:011|5[0-9]{2})[0-9]{12}`))

	// JCBRedactor is a Redactor that will redact JCB Numbers
	JCBRedactor = core.Must(NewRedactor(`(?:2131|1800|35\d{3})\d{11}`))

	// MasterCardRedactor is a Redactor that will redact MasterCard Numbers
	MasterCardRedactor = core.Must(NewRedactor(`(5[1-5][0-9]{2}|222[1-9]|22[3-9][0-9]|2[3-6][0-9]{2}|27[01][0-9]|2720)[0-9]{12}`))

	// VISARedactor is a Redactor that will redact VISA Credit Card Numbers
	VISARedactor = core.Must(NewRedactor(`4[0-9]{12}(?:[0-9]{3})?`))

	// PhoneRedactor is a Redactor that will redact Phone Numbers
	PhoneRedactor = core.Must(NewRedactor(`(\+?[0-9]{1,3}[-. ]?[0-9]{3}[-. ]?[0-9]{3}[-. ]?[0-9]{4})|(\+?[0-9]{1,3}[-. ]?[0-9.\-]+)`))

	// EmailRedactor is a Redactor that will redact Email Addresses
	EmailRedactor = core.Must(NewRedactor(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,4}`))

	// IPRedactor is a Redactor that will redact IP Addresses (v4 and v6)
	IPRedactor = core.Must(NewRedactor(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})|([0-9a-fA-F:]{2,39})`))
	// IPV4Redactor is a Redactor that will redact IP v4 Addresses
	IPV4Redactor = core.Must(NewRedactor(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`))

	// IPV6Redactor is a Redactor that will redact IP v6 Addresses
	IPV6Redactor = core.Must(NewRedactor(`[0-9a-fA-F:]{2,39}`))

	// MACRedactor is a Redactor that will redact MAC Addresses
	MACRedactor = core.Must(NewRedactor(`([0-9A-Fa-f]{2}[:-]){5}([0-9A-Fa-f]{2})`))
)

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
