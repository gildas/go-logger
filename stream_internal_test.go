package logger

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateUnbufferedStreamWithDEBUG(t *testing.T) {
	os.Setenv("DEBUG", "1")
	defer os.Unsetenv("DEBUG")
	stream := CreateStreamWithDestination()
	assert.True(t, stream.(*StdoutStream).Unbuffered, "Stdout should be unbuffered with DEBUG set")
}
