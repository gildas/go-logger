package logger_test

import (
	"strings"

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

func (suite *LoggerSuite) TestShouldNotRedactMessageWithNoMatch() {
	redactor := core.Must(logger.NewRedactor(`\+[0-9]{23}`))
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
		"msg":      `message with sensitive \(\+13178723000\) data`,
		"name":     "test",
		"pid":      "[0-9]+",
		"scope":    "main",
		"tid":      "[0-9]+",
		"time":     `[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+Z`,
		"topic":    "main",
		"v":        "0",
	})
}

func (suite *LoggerSuite) TestCanRedactMessageWithSeveralRedactors() {
	output := CaptureStdout(func() {
		log := logger.Create(
			"test",
			&logger.StdoutStream{Unbuffered: true},
			logger.VISARedactor,
			logger.AMEXRedactor,
		)
		log.Infof("message with sensitive (VISA: 4977949494949497) data")
		log.Infof("message with sensitive (AMEX: 370000000100018) data")
	})
	lines := strings.Split(output, "\n")[0:2]
	suite.LogLineEqual(lines[0], map[string]string{
		"hostname": `[a-zA-Z_0-9\-\.]+`,
		"level":    "30",
		"msg":      `message with sensitive \(VISA: REDACTED\) data`,
		"name":     "test",
		"pid":      "[0-9]+",
		"scope":    "main",
		"tid":      "[0-9]+",
		"time":     `[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+Z`,
		"topic":    "main",
		"v":        "0",
	})
	suite.LogLineEqual(lines[1], map[string]string{
		"hostname": `[a-zA-Z_0-9\-\.]+`,
		"level":    "30",
		"msg":      `message with sensitive \(AMEX: REDACTED\) data`,
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

func (suite *LoggerSuite) TestCanRedactCreditCardCard() {
	redactor := logger.CreditCardRedactor
	redacted, ok := redactor.Redact("message with 30569309025904")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with nothing")
	suite.Assert().Falsef(ok, "Redactor %T should not have matched", redactor)
	suite.Assert().Equal("message with nothing", redacted)
}

func (suite *LoggerSuite) TestCanRedactAMEXCard() {
	redactor := logger.AMEXRedactor
	redacted, ok := redactor.Redact("message with 370000000100018")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with nothing")
	suite.Assert().Falsef(ok, "Redactor %T should not have matched", redactor)
	suite.Assert().Equal("message with nothing", redacted)
}

func (suite *LoggerSuite) TestCanRedactVISACard() {
	redactor := logger.VISARedactor
	redacted, ok := redactor.Redact("message with 4111111111111111")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with nothing")
	suite.Assert().Falsef(ok, "Redactor %T should not have matched", redactor)
	suite.Assert().Equal("message with nothing", redacted)
}

func (suite *LoggerSuite) TestCanRedactDinersClubCard() {
	redactor := logger.DinersClubRedactor
	redacted, ok := redactor.Redact("message with 30569309025904")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with nothing")
	suite.Assert().Falsef(ok, "Redactor %T should not have matched", redactor)
	suite.Assert().Equal("message with nothing", redacted)
}

func (suite *LoggerSuite) TestCanRedactDiscoverCard() {
	redactor := logger.DiscoverRedactor
	redacted, ok := redactor.Redact("message with 6011000990139424")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with nothing")
	suite.Assert().Falsef(ok, "Redactor %T should not have matched", redactor)
	suite.Assert().Equal("message with nothing", redacted)
}

func (suite *LoggerSuite) TestCanRedactMasterCardCard() {
	redactor := logger.MasterCardRedactor
	redacted, ok := redactor.Redact("message with 5105105105105100")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with nothing")
	suite.Assert().Falsef(ok, "Redactor %T should not have matched", redactor)
	suite.Assert().Equal("message with nothing", redacted)
}

func (suite *LoggerSuite) TestCanRedactPhone() {
	redactor := logger.PhoneRedactor
	redacted, ok := redactor.Redact("message with +13178723000")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with +8112345678")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with nothing")
	suite.Assert().Falsef(ok, "Redactor %T should not have matched", redactor)
	suite.Assert().Equal("message with nothing", redacted)
}

func (suite *LoggerSuite) TestCanRedactEmail() {
	redactor := logger.EmailRedactor
	redacted, ok := redactor.Redact("message with john.doe@acme.com")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with nothing")
	suite.Assert().Falsef(ok, "Redactor %T should not have matched", redactor)
	suite.Assert().Equal("message with nothing", redacted)
}

func (suite *LoggerSuite) TestCanRedactIP() {
	redactor := logger.IPRedactor
	redacted, ok := redactor.Redact("message with 192.168.1.1")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with 2001:db8:3333:4444:5555:6666:7777:8888")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with nothing")
	suite.Assert().Falsef(ok, "Redactor %T should not have matched", redactor)
	suite.Assert().Equal("message with nothing", redacted)
}

func (suite *LoggerSuite) TestCanRedactIPV4() {
	redactor := logger.IPV4Redactor
	redacted, ok := redactor.Redact("message with 192.168.1.1")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with nothing")
	suite.Assert().Falsef(ok, "Redactor %T should not have matched", redactor)
	suite.Assert().Equal("message with nothing", redacted)
}

func (suite *LoggerSuite) TestCanRedactIPV6() {
	redactor := logger.IPV6Redactor
	redacted, ok := redactor.Redact("message with 2600:::")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with 2001:db8:3333:4444:5555:6666:7777:8888")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	// redacted, ok = redactor.Redact("message with 2001:db8:3333:4444:5555:6666:1.2.3.4")
	// suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	// suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with nothing")
	suite.Assert().Falsef(ok, "Redactor %T should not have matched", redactor)
	suite.Assert().Equal("message with nothing", redacted)
}

func (suite *LoggerSuite) TestCanRedactMAC() {
	redactor := logger.MACRedactor
	redacted, ok := redactor.Redact("message with 2C:54:91:88:C9:E3")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with nothing")
	suite.Assert().Falsef(ok, "Redactor %T should not have matched", redactor)
	suite.Assert().Equal("message with nothing", redacted)
}

func (suite *LoggerSuite) TestCanRedactWithKeysToRedact() {
	metadata := Metadata{"12345678", "Taro Yamamoto", "Tokyo"}
	output := CaptureStdout(func() {
		log := logger.Create("test", &logger.StdoutStream{Unbuffered: true})
		log.RecordWithKeysToRedact("metadata", metadata, "name", "city").Infof("message")
	})
	suite.LogLineEqual(output, map[string]string{
		"metadata": `map\[city:REDACTED name:REDACTED userId:12345678\]`,
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
