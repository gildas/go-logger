//go:build linux
// +build linux

package logger

import (
	"syscall"
)

// gettid returns the current thread identifier
func gettid() int { return syscall.Gettid() }
