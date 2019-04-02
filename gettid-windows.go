// +build windows

package logger

import ()

// Gettid returns the current thread identifier
func Gettid() int { return 0 }
