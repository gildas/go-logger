package logger

import (
	"crypto/sha256"
	"fmt"
)

func Redact(value interface{}) string {
	if value == nil {
		return ""
	}
	if len(fmt.Sprintf("%s", value)) > 0 {
		return "REDACTED"
	}
	return ""
}

func RedactWithHash(value interface{}) string {
	return RedactWithPrefixedHash("REDACTED", value)
}

func RedactWithPrefixedHash(prefix string, value interface{}) string {
	if value == nil {
		return ""
	}
	redacted := bufferPool.Get()
	defer bufferPool.Put(redacted)

	redacted.WriteString(prefix)
	redacted.WriteString("-")
	redacted.WriteString(fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s", value))))[:10])
	return redacted.String()
}
