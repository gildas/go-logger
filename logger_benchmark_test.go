package logger_test

import (
	gologger "log"
	"testing"

	"github.com/gildas/go-logger"
)

func BenchmarkLogger_standard(b *testing.B) {
	file, teardown := CreateTempFile()
	defer teardown()
	log := gologger.New(file, "", gologger.LstdFlags)

	for i := 0; i < b.N; i++ {
		log.Printf("Hello World! (%d)", i)
	}
}

func BenchmarkLogger_buffered(b *testing.B) {
	file, teardown := CreateTempFile()
	defer teardown()
	log := logger.Create("benchmark", &logger.FileStream{Path: file.Name()})

	for i := 0; i < b.N; i++ {
		log.Infof("Hello World! (%d)", i)
	}
	log.Flush()
}

func BenchmarkLogger_unbuffered(b *testing.B) {
	file, teardown := CreateTempFile()
	defer teardown()
	log := logger.Create("benchmark", &logger.FileStream{Path: file.Name(), Unbuffered: true})

	for i := 0; i < b.N; i++ {
		log.Infof("Hello World! (%d)", i)
	}
}

func BenchmarkLogger_buffered_with_marshal(b *testing.B) {
	file, teardown := CreateTempFile()
	defer teardown()
	log := logger.Create("benchmark", &logger.FileStream{Path: file.Name()})

	object := struct {
		Name string
		Age  int
	}{
		Name: "John Doe",
		Age:  42,
	}
	for i := 0; i < b.N; i++ {
		log.Record("object", object).Infof("Hello World! (%d)", i)
	}
	log.Flush()
}

func BenchmarkLogger_buffered_above_level(b *testing.B) {
	file, teardown := CreateTempFile()
	defer teardown()
	log := logger.Create("benchmark", &logger.FileStream{Path: file.Name()})

	for i := 0; i < b.N; i++ {
		log.Tracef("Hello World! (%d)", i)
	}
	log.Flush()
}
