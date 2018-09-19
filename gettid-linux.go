// +build linux

package logger

import (
	"syscall"
)

// Gettid returns the current thread identifier
func Gettid() int { return syscall.Gettid() }
