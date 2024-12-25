package logger

import (
	"crypto/sha256"
	"fmt"
)

// Redact redacts a stringable value
func Redact(value interface{}) string {
	if value == nil {
		return ""
	}
	if len(fmt.Sprintf("%s", value)) > 0 {
		return "REDACTED"
	}
	return ""
}

// RedactSlice redacts all items in a slice of redactable items
func RedactSlice[T Redactable](items []T) []any {
	redacted := make([]any, len(items))
	for i, item := range items {
		redacted[i] = item.Redact()
	}
	return redacted
}

func RedactMap[S comparable, T Redactable](items map[S]T) map[S]any {
	redacted := make(map[S]any, len(items))
	for key, item := range items {
		redacted[key] = item.Redact()
	}
	return redacted
}

// RedactWithHash redacts a value with a hash
func RedactWithHash(value interface{}) string {
	return RedactWithPrefixedHash("REDACTED", value)
}

// RedactWithPrefixedHash redacts a value with a prefix and a hash
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
