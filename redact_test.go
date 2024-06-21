package logger_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/gildas/go-core"
	"github.com/gildas/go-logger"
	"github.com/stretchr/testify/suite"
)

type RedactSuite struct {
	LoggerSuite
	Name string
}

func TestRedactSuite(t *testing.T) {
	suite.Run(t, new(RedactSuite))
}

func (suite *RedactSuite) SetupSuite() {
	suite.Name = strings.TrimSuffix(reflect.TypeOf(suite).Elem().Name(), "Suite")
}

func (suite *RedactSuite) TestCanRedactSensitiveStruct() {
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

func (suite *RedactSuite) TestCanRedactMessage() {
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

func (suite *RedactSuite) TestShouldNotRedactMessageWithNoMatch() {
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

func (suite *RedactSuite) TestCanRedactMessageWithSeveralRedactors() {
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

func (suite *RedactSuite) TestCanRedactAsString() {
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

func (suite *RedactSuite) TestCanRedactCreditCardCard() {
	redactor := logger.CreditCardRedactor
	redacted, ok := redactor.Redact("message with 30569309025904")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with 3700 000001 00018")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with 4111 1111 1111 1111")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with 3056 9309 0259 04")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with 6011 0009 9013 9424")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with 5105 1051 0510 5100")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with 3569 9900 1009 5841")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with nothing")
	suite.Assert().Falsef(ok, "Redactor %T should not have matched", redactor)
	suite.Assert().Equal("message with nothing", redacted)
}

func (suite *RedactSuite) TestCanRedactAMEXCard() {
	redactor := logger.AMEXRedactor
	redacted, ok := redactor.Redact("message with 370000000100018")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with 3700 000001 00018")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with nothing")
	suite.Assert().Falsef(ok, "Redactor %T should not have matched", redactor)
	suite.Assert().Equal("message with nothing", redacted)
}

func (suite *RedactSuite) TestCanRedactVISACard() {
	redactor := logger.VISARedactor
	redacted, ok := redactor.Redact("message with 4111111111111111")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with 4111 1111 1111 1111")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with nothing")
	suite.Assert().Falsef(ok, "Redactor %T should not have matched", redactor)
	suite.Assert().Equal("message with nothing", redacted)
}

func (suite *RedactSuite) TestCanRedactDinersClubCard() {
	redactor := logger.DinersClubRedactor
	redacted, ok := redactor.Redact("message with 30569309025904")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with 3056 9309 0259 04")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with nothing")
	suite.Assert().Falsef(ok, "Redactor %T should not have matched", redactor)
	suite.Assert().Equal("message with nothing", redacted)
}

func (suite *RedactSuite) TestCanRedactDiscoverCard() {
	redactor := logger.DiscoverRedactor
	redacted, ok := redactor.Redact("message with 6011000990139424")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with 6011 0009 9013 9424")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with nothing")
	suite.Assert().Falsef(ok, "Redactor %T should not have matched", redactor)
	suite.Assert().Equal("message with nothing", redacted)
}

func (suite *RedactSuite) TestCanRedactMasterCardCard() {
	redactor := logger.MasterCardRedactor
	redacted, ok := redactor.Redact("message with 5105105105105100")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with 5105 1051 0510 5100")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with nothing")
	suite.Assert().Falsef(ok, "Redactor %T should not have matched", redactor)
	suite.Assert().Equal("message with nothing", redacted)
}

func (suite *RedactSuite) TestCanRedactJCBCard() {
	redactor := logger.JCBRedactor
	redacted, ok := redactor.Redact("message with 3569990010095841")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with 3569 9900 1009 5841")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with nothing")
	suite.Assert().Falsef(ok, "Redactor %T should not have matched", redactor)
	suite.Assert().Equal("message with nothing", redacted)
}

func (suite *RedactSuite) TestCanRedactPhone() {
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

func (suite *RedactSuite) TestCanRedactEmail() {
	redactor := logger.EmailRedactor
	redacted, ok := redactor.Redact("message with john.doe@acme.com")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with nothing")
	suite.Assert().Falsef(ok, "Redactor %T should not have matched", redactor)
	suite.Assert().Equal("message with nothing", redacted)
}

func (suite *RedactSuite) TestCanRedactIP() {
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

func (suite *RedactSuite) TestCanRedactIPV4() {
	redactor := logger.IPV4Redactor
	redacted, ok := redactor.Redact("message with 192.168.1.1")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with nothing")
	suite.Assert().Falsef(ok, "Redactor %T should not have matched", redactor)
	suite.Assert().Equal("message with nothing", redacted)
}

func (suite *RedactSuite) TestCanRedactIPV6() {
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

func (suite *RedactSuite) TestCanRedactMAC() {
	redactor := logger.MACRedactor
	redacted, ok := redactor.Redact("message with 2C:54:91:88:C9:E3")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with nothing")
	suite.Assert().Falsef(ok, "Redactor %T should not have matched", redactor)
	suite.Assert().Equal("message with nothing", redacted)
}

func (suite *RedactSuite) TestCanMergeEmpty() {
	suite.Assert().Equal(logger.VISARedactor.String(), logger.VISARedactor.Merge().String())
}

func (suite *RedactSuite) TestCanMergeWithRedactor() {
	expected := `3[47]\d{2}[- ]*\d{6}[- ]*\d{5}|4\d{3}[- ]*\d{4}[- ]*\d{4}[- ]*\d{4}`
	redactor := logger.AMEXRedactor.Merge(logger.VISARedactor)
	suite.Assert().Equal(expected, redactor.String())
	redacted, ok := redactor.Redact("message with 3700 000001 00018")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with 4111 1111 1111 1111")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with 5105 1051 0510 5100")
	suite.Assert().Falsef(ok, "Redactor %T should not have matched", redactor)
	suite.Assert().Equal("message with 5105 1051 0510 5100", redacted)
	redacted, ok = redactor.Redact("message with nothing")
	suite.Assert().Falsef(ok, "Redactor %T should not have matched", redactor)
	suite.Assert().Equal("message with nothing", redacted)
}

func (suite *RedactSuite) TestCanMergeWithRedactors() {
	expected := `3[47]\d{2}[- ]*\d{6}[- ]*\d{5}|4\d{3}[- ]*\d{4}[- ]*\d{4}[- ]*\d{4}|(?:5[1-5][0-9]{2}|222[1-9]|22[3-9][0-9]|2[3-6][0-9]{2}|27[01][0-9]|2720)[- ]*\d{4}[- ]*\d{4}[- ]*\d{4}`
	redactor := logger.AMEXRedactor.Merge(logger.VISARedactor, *logger.MasterCardRedactor)
	suite.Assert().Equal(expected, redactor.String())
	redacted, ok := redactor.Redact("message with 3700 000001 00018")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with 4111 1111 1111 1111")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with 5105 1051 0510 5100")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with nothing")
	suite.Assert().Falsef(ok, "Redactor %T should not have matched", redactor)
	suite.Assert().Equal("message with nothing", redacted)
}

func (suite *RedactSuite) TestCanMergeWithPipedRedactors() {
	expected := `3[47]\d{2}[- ]*\d{6}[- ]*\d{5}|4\d{3}[- ]*\d{4}[- ]*\d{4}[- ]*\d{4}|(?:5[1-5][0-9]{2}|222[1-9]|22[3-9][0-9]|2[3-6][0-9]{2}|27[01][0-9]|2720)[- ]*\d{4}[- ]*\d{4}[- ]*\d{4}`
	redactor := logger.AMEXRedactor.Merge(logger.VISARedactor).Merge(logger.MasterCardRedactor)
	suite.Assert().Equal(expected, redactor.String())
	redacted, ok := redactor.Redact("message with 3700 000001 00018")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with 4111 1111 1111 1111")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with 5105 1051 0510 5100")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with nothing")
	suite.Assert().Falsef(ok, "Redactor %T should not have matched", redactor)
	suite.Assert().Equal("message with nothing", redacted)
}

func (suite *RedactSuite) TestCanMergeWithString() {
	expected := `3[47]\d{2}[- ]*\d{6}[- ]*\d{5}|4\d{3}[- ]*\d{4}[- ]*\d{4}[- ]*\d{4}`
	redactor := logger.AMEXRedactor.Merge(logger.VISARedactor.String())
	suite.Assert().Equal(expected, redactor.String())
	redacted, ok := redactor.Redact("message with 3700 000001 00018")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with 4111 1111 1111 1111")
	suite.Assert().Truef(ok, "Redactor %T should have matched", redactor)
	suite.Assert().Equal("message with REDACTED", redacted)
	redacted, ok = redactor.Redact("message with nothing")
	suite.Assert().Falsef(ok, "Redactor %T should not have matched", redactor)
	suite.Assert().Equal("message with nothing", redacted)
}

type stringer logger.Redactor

func (s stringer) String() string {
	return logger.Redactor(s).String()
}

func (suite *RedactSuite) TestCanMergeWithStringer() {
	expected := `3[47]\d{2}[- ]*\d{6}[- ]*\d{5}|4\d{3}[- ]*\d{4}[- ]*\d{4}[- ]*\d{4}`
	redactor := logger.AMEXRedactor.Merge(stringer(*logger.VISARedactor))
	suite.Assert().Equal(expected, redactor.String())
}

func (suite *RedactSuite) TestMergeShouldPanicWithWrongType() {
	suite.Require().Panics(func() {
		_ = logger.VISARedactor.Merge(42)
	})
}

func (suite *RedactSuite) TestCanUseSeveralRedactors() {
	output := CaptureStdout(func() {
		mainLog := logger.Create(
			"test",
			&logger.StdoutStream{Unbuffered: true},
		)
		log := mainLog.Child(
			nil,
			nil,
			logger.PhoneRedactor,
			logger.EmailRedactor,
		)
		log.Infof("message with sensitive (+13178723000) data")
		log.Infof("message with sensitive (john.doe@acme.com) data")
	})
	lines := strings.Split(output, "\n")[0:2]
	suite.LogLineEqual(lines[0], map[string]string{
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
	suite.LogLineEqual(lines[1], map[string]string{
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

func (suite *RedactSuite) TestCanRedactWithKeysToRedact() {
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
