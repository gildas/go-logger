package logger

import (
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"regexp"
)

var unobfuscateRex = regexp.MustCompile("(.*)!ENC!:{(.*)}(.*)")

// Obfuscate obfuscates the given string
func (l *Logger) Obfuscate(value string) string {
	if l.obfuscationKey != nil {
		var err error
		var gcm cipher.AEAD

		if gcm, err = cipher.NewGCM(l.obfuscationKey); err == nil {
			nonce := make([]byte, gcm.NonceSize())
			if _, err = io.ReadFull(rand.Reader, nonce); err == nil {
				return "!ENC!:{" + base64.URLEncoding.EncodeToString(gcm.Seal(nonce, nonce, []byte(value), nil)) + "}"
			}
		}
		l.Child("logger", "obfuscate").Errorf("Failed to obfuscate", err)
		return value
	}
	l.Child("logger", "obfuscate").Warnf("WARNING, the next string will not be obfuscated because no obfuscation key was provided")
	return value
}

// Unobfuscate the given string
func (l *Logger) Unobfuscate(value string) (unobfuscated string, err error) {
	if l.obfuscationKey != nil {
		if components := unobfuscateRex.FindStringSubmatch(value); len(components) == 4 {
			var decoded []byte

			if decoded, err = base64.URLEncoding.DecodeString(components[2]); err == nil {
				var gcm cipher.AEAD

				if gcm, err = cipher.NewGCM(l.obfuscationKey); err == nil {
					var decrypted []byte

					nonce := decoded[:gcm.NonceSize()]
					decoded = decoded[gcm.NonceSize():]
					if decrypted, err = gcm.Open(nil, nonce, decoded, nil); err == nil {
						return components[1] + string(decrypted) + components[3], nil
					}
				}
			}
			return value, err
		}
	}
	return value, nil
}
