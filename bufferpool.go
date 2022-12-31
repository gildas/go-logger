package logger

import (
	"github.com/oxtoacart/bpool"
)

var bufferPool = bpool.NewBufferPool(64)