//go:build linux || darwin
// +build linux darwin

package main

import (
	"crypto/aes"

	"github.com/gildas/go-core"
	"github.com/gildas/go-logger"
)

func main() {
	cipherBlock, err := aes.NewCipher([]byte(core.GetEnvAsString("CIPHER_KEY", "1234567890123456")))
	if err != nil {
		panic(err)
	}
	log := logger.Create("sample", &logger.StdoutStream{Unbuffered: true}, cipherBlock)

	log.Infof("This following text should be hidden: %s and this text should not", log.Obfuscate("secret stuff"))
}
