//go:build windows
// +build windows

package logger

import ()

// gettid returns the current thread identifier
func gettid() int { return 0 }
