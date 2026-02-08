package logger_test

import (
	"crypto/aes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/gildas/go-core"
	"github.com/gildas/go-logger"
	"github.com/stretchr/testify/suite"
)

type ObfuscatorSuite struct {
	LoggerSuite
	Name      string
	CipherKey string
}

func TestObfuscatorSuite(t *testing.T) {
	suite.Run(t, new(ObfuscatorSuite))
}

func (suite *ObfuscatorSuite) SetupSuite() {
	suite.Name = strings.TrimSuffix(reflect.TypeOf(suite).Elem().Name(), "Suite")
	suite.CipherKey = core.GetEnvAsString("CIPHER_KEY", "1234567890123456")
}

func (suite *ObfuscatorSuite) TearDownSuite() {
	if suite.T().Failed() {
		suite.T().Log("At least one test failed, we are not cleaning")
	}
}

func (suite *ObfuscatorSuite) TestCanObfuscate() {
	cipherBlock, err := aes.NewCipher([]byte(suite.CipherKey))
	suite.Require().NoError(err, "Failed to create cipher block")

	customer := User{"12345678", "John Doe", nil}
	output := CaptureStdout(func() {
		log := logger.Create("test", cipherBlock, &logger.StdoutStream{Unbuffered: true})
		log.Record("customer", customer).Infof("This should be hidden: %s and more text", log.Obfuscate("secret stuff"))
	})
	// suite.T().Logf("output:\n%s", output)
	lines := strings.Split(output, "\n")
	lines = lines[0 : len(lines)-1] // remove the last empty line
	suite.Require().Len(lines, 1, "There should be 1 line in the log output, found %d", len(lines))
	suite.LogLineEqual(lines[0], map[string]string{
		"customer": `map\[id:12345678 name:REDACTED\]`,
		"hostname": `[a-zA-Z_0-9\-\.]+`,
		"level":    "30",
		"msg":      "This should be hidden: !ENC!:{.*} and more text",
		"name":     "test",
		"pid":      "[0-9]+",
		"scope":    "main",
		"tid":      "[0-9]+",
		"time":     `[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+Z`,
		"topic":    "main",
		"v":        "0",
	})

	log := logger.Create("test", cipherBlock, &logger.StdoutStream{Unbuffered: true})
	var entry map[string]any
	err = json.Unmarshal([]byte(lines[0]), &entry)
	suite.Require().NoError(err, "Failed to unmarshal log line #0")
	cleartext, err := log.Unobfuscate(entry["msg"].(string))
	suite.Require().NoError(err, "Failed to unobfuscate log line #0")
	suite.Assert().Equal("This should be hidden: secret stuff and more text", cleartext)
}

func (suite *ObfuscatorSuite) TestCanObfuscateWithoutKey() {
	// In that case, the logger will not obfuscate anything and should add a warning in the logs
	customer := User{"12345678", "John Doe", nil}
	output := CaptureStdout(func() {
		log := logger.Create("test", &logger.StdoutStream{Unbuffered: true})
		log.Record("customer", customer).Infof("%s", log.Obfuscate("secret stuff"))
	})
	lines := strings.Split(output, "\n")
	lines = lines[0 : len(lines)-1] // remove the last empty line
	suite.Require().Len(lines, 2, "There should be 2 lines in the log output, found %d", len(lines))
	suite.LogLineEqual(lines[0], map[string]string{
		"hostname": `[a-zA-Z_0-9\-\.]+`,
		"level":    "40",
		"msg":      "WARNING, the next string will not be obfuscated because no obfuscation key was provided",
		"name":     "test",
		"pid":      "[0-9]+",
		"scope":    "obfuscate",
		"tid":      "[0-9]+",
		"time":     `[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+Z`,
		"topic":    "logger",
		"v":        "0",
	})
	suite.LogLineEqual(lines[1], map[string]string{
		"customer": `map\[id:12345678 name:REDACTED\]`,
		"hostname": `[a-zA-Z_0-9\-\.]+`,
		"level":    "30",
		"msg":      "secret stuff",
		"name":     "test",
		"pid":      "[0-9]+",
		"scope":    "main",
		"tid":      "[0-9]+",
		"time":     `[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+Z`,
		"topic":    "main",
		"v":        "0",
	})
}

func (suite *ObfuscatorSuite) TestCanUnobfuscate() {
	cipherBlock, err := aes.NewCipher([]byte(suite.CipherKey))
	suite.Require().NoError(err, "Failed to create cipher block")
	log := logger.Create("test", cipherBlock, &logger.StdoutStream{Unbuffered: true})

	obfuscated := "This should be hidden: !ENC!:{rAdNoayKlX5t8uKvmZ8FBadpq--P2bbem7Dqero33wL3gguQcHKiBA==} and more text"
	cleartext, err := log.Unobfuscate(obfuscated)
	suite.Require().NoError(err, "Failed to unobfuscate log line #0")
	suite.Assert().Equal("This should be hidden: secret stuff and more text", cleartext)

	obfuscated = "!ENC!:{rAdNoayKlX5t8uKvmZ8FBadpq--P2bbem7Dqero33wL3gguQcHKiBA==} and more text"
	cleartext, err = log.Unobfuscate(obfuscated)
	suite.Require().NoError(err, "Failed to unobfuscate log line #0")
	suite.Assert().Equal("secret stuff and more text", cleartext)

	obfuscated = "This should be hidden: !ENC!:{rAdNoayKlX5t8uKvmZ8FBadpq--P2bbem7Dqero33wL3gguQcHKiBA==}"
	cleartext, err = log.Unobfuscate(obfuscated)
	suite.Require().NoError(err, "Failed to unobfuscate log line #0")
	suite.Assert().Equal("This should be hidden: secret stuff", cleartext)

	obfuscated = "!ENC!:{rAdNoayKlX5t8uKvmZ8FBadpq--P2bbem7Dqero33wL3gguQcHKiBA==}"
	cleartext, err = log.Unobfuscate(obfuscated)
	suite.Require().NoError(err, "Failed to unobfuscate log line #0")
	suite.Assert().Equal("secret stuff", cleartext)

	obfuscated = "some text"
	cleartext, err = log.Unobfuscate(obfuscated)
	suite.Require().NoError(err, "Failed to unobfuscate log line #0")
	suite.Assert().Equal("some text", cleartext)

}

type bogusCipherBlock struct{}

func (block bogusCipherBlock) BlockSize() int          { return 0 }
func (block bogusCipherBlock) Encrypt(dst, src []byte) {}
func (block bogusCipherBlock) Decrypt(dst, src []byte) {}

func (suite *ObfuscatorSuite) TestShouldFailObfuscateWithWrongBlockSize() {
	output := CaptureStdout(func() {
		log := logger.Create("test", bogusCipherBlock{}, &logger.StdoutStream{Unbuffered: true})
		log.Infof("%s", log.Obfuscate("secret stuff"))
	})
	lines := strings.Split(output, "\n")
	lines = lines[0 : len(lines)-1] // remove the last empty line
	suite.Require().Len(lines, 2, "There should be 2 lines in the log output, found %d", len(lines))
	suite.LogLineEqual(lines[0], map[string]string{
		"err":      `map\[\]`,
		"hostname": `[a-zA-Z_0-9\-\.]+`,
		"level":    "50",
		"msg":      "Failed to obfuscate. Error:",
		"name":     "test",
		"pid":      "[0-9]+",
		"scope":    "obfuscate",
		"tid":      "[0-9]+",
		"time":     `[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+Z`,
		"topic":    "logger",
		"v":        "0",
	})
	suite.LogLineEqual(lines[1], map[string]string{
		"hostname": `[a-zA-Z_0-9\-\.]+`,
		"level":    "30",
		"msg":      "secret stuff",
		"name":     "test",
		"pid":      "[0-9]+",
		"scope":    "main",
		"tid":      "[0-9]+",
		"time":     `[0-9]+-[0-9]+-[0-9]+T[0-9]+:[0-9]+:[0-9]+Z`,
		"topic":    "main",
		"v":        "0",
	})
}

func (suite *ObfuscatorSuite) TestShouldFailUnobfuscateWithBogusEncryptedData() {
	log := logger.Create("test", bogusCipherBlock{}, &logger.StdoutStream{Unbuffered: true})
	_, err := log.Unobfuscate("!ENC!:{dummyvalue}")
	suite.Require().Error(err, "Should Fail to unobfuscate log line #0")
}

func (suite *ObfuscatorSuite) TestShouldFailUnobfuscateWithWrongBlockSize() {
	log := logger.Create("test", bogusCipherBlock{}, &logger.StdoutStream{Unbuffered: true})
	_, err := log.Unobfuscate("!ENC!:{K8KCYXK_vo8VkJhECFel4_qhi1EXlKk0Hd7Um2hiAeAKWkgiZWUvCA==}")
	suite.Require().Error(err, "Should Fail to unobfuscate log line #0")
}
