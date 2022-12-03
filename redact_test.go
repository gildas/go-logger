package logger_test

import (
	"github.com/gildas/go-core"
	"github.com/gildas/go-logger"
)

func (suite *LoggerSuite) TestCanRedactSensitiveStruct() {
	customer := User{"12345678", "John Doe", nil}
	output := CaptureStdout(func() {
		log := logger.Create("test", &logger.StdoutStream{Unbuffered: true})
		log.Record("customer", customer).Infof("message")
	})
	suite.LogLineEqual(output, map[string]string{
		"customer": `map\[id:12345678 name:REDACTED\]`,
		"hostname": `[a-zA-Z_0-9\-\.]+`,
		"level":    "30",
		"msg":      "message",
		"name":     "test",
		"pid":      "[0-9]+",
		"scope":    "main",
		"tid":      "[0-9]+",
		"time":     `[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+Z`,
		"topic":    "main",
		"v":        "0",
	})
}

func (suite *LoggerSuite) TestCanRedactMessage() {
	redactor := core.Must(logger.NewRedactor(`\+[0-9]{11}`))
	suite.Require().NotEmpty(redactor.String())
	output := CaptureStdout(func() {
		log := logger.Create(
			"test",
			&logger.StdoutStream{Unbuffered: true},
			redactor,
		)
		log.Infof("message with sensitive (+13178723000) data")
	})
	suite.LogLineEqual(output, map[string]string{
		"hostname": `[a-zA-Z_0-9\-\.]+`,
		"level":    "30",
		"msg":      `message with sensitive \(REDACTED\) data`,
		"name":     "test",
		"pid":      "[0-9]+",
		"scope":    "main",
		"tid":      "[0-9]+",
		"time":     `[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+Z`,
		"topic":    "main",
		"v":        "0",
	})
}

func (suite *LoggerSuite) TestCanRedactAsString() {
	suite.Assert().Equal("", logger.Redact(nil))
	suite.Assert().Equal("", logger.Redact(""))
	redacted := logger.Redact("John Doe")
	suite.Assert().Equal("REDACTED", redacted)
	redacted = logger.Redact(User{"12345678", "John Doe", nil})
	suite.Assert().Equal("REDACTED", redacted)
	suite.Assert().Equal("", logger.RedactWithHash(nil))
	suite.Assert().Equal("REDACTED-6cea57c2fb", logger.RedactWithHash("John Doe"))
	suite.Assert().Equal("Name-6cea57c2fb", logger.RedactWithPrefixedHash("Name", "John Doe"))
}

