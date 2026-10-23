package logger_test

import (
	"crypto/aes"
	gologger "log"
	"testing"

	"github.com/gildas/go-logger"
)

var benchmarkJSONSink []byte

type benchmarkStream struct {
	converter logger.Converter
}

func (stream *benchmarkStream) Write(record *logger.Record) error {
	var err error
	benchmarkJSONSink, err = stream.converter.Convert(record).MarshalJSON()
	return err
}

func (stream *benchmarkStream) ShouldWrite(logger.Level, string, string) bool { return true }
func (stream *benchmarkStream) ShouldLogSourceInfo() bool                     { return false }
func (stream *benchmarkStream) Flush()                                        {}
func (stream *benchmarkStream) Close()                                        {}
func (stream *benchmarkStream) GetFilterLevels() logger.LevelSet {
	return logger.NewLevelSet(logger.TRACE)
}
func (stream *benchmarkStream) Clone() logger.Streamer {
	return &benchmarkStream{converter: &logger.BunyanConverter{}}
}

type benchmarkJSONRecord struct {
	UserID  string `json:"userId"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Active  bool   `json:"active"`
	Retries int    `json:"retries"`
}

func (record benchmarkJSONRecord) Redact(keys ...string) any {
	for _, key := range keys {
		switch key {
		case "name":
			record.Name = logger.Redact(record.Name)
		case "email":
			record.Email = logger.Redact(record.Email)
		}
	}
	return record
}

func newBenchmarkLogger(b *testing.B, options ...any) *logger.Logger {
	b.Helper()
	parameters := append([]any{&benchmarkStream{converter: &logger.BunyanConverter{}}}, options...)
	log := logger.Create("benchmark", parameters...)
	b.ReportAllocs()
	return log
}

func BenchmarkLogger_standard(b *testing.B) {
	file, teardown := CreateTempFile()
	defer teardown()
	log := gologger.New(file, "", gologger.LstdFlags)

	for i := 0; b.Loop(); i++ {
		log.Printf("Hello World! (%d)", i)
	}
}

func BenchmarkLogger_buffered(b *testing.B) {
	file, teardown := CreateTempFile()
	defer teardown()
	log := logger.Create("benchmark", &logger.FileStream{Path: file.Name()})

	for i := 0; b.Loop(); i++ {
		log.Infof("Hello World! (%d)", i)
	}
	log.Flush()
}

func BenchmarkLogger_unbuffered(b *testing.B) {
	file, teardown := CreateTempFile()
	defer teardown()
	log := logger.Create("benchmark", &logger.FileStream{Path: file.Name(), Unbuffered: true})

	for i := 0; b.Loop(); i++ {
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
	for i := 0; b.Loop(); i++ {
		log.Record("object", object).Infof("Hello World! (%d)", i)
	}
	log.Flush()
}

func BenchmarkLogger_buffered_above_level(b *testing.B) {
	file, teardown := CreateTempFile()
	defer teardown()
	log := logger.Create("benchmark", &logger.FileStream{Path: file.Name()})

	for i := 0; b.Loop(); i++ {
		log.Tracef("Hello World! (%d)", i)
	}
	log.Flush()
}

func BenchmarkLogger_simple_message(b *testing.B) {
	log := newBenchmarkLogger(b)
	b.ResetTimer()

	for i := 0; b.Loop(); i++ {
		log.Infof("Request completed")
	}
}

func BenchmarkLogger_complex_message(b *testing.B) {
	log := newBenchmarkLogger(b)
	b.ResetTimer()

	for i := 0; b.Loop(); i++ {
		log.Infof("Request %s for user %d completed in %dms (cached=%t, result=%v)", "/api/v1/items", 12345, 27, true, []string{"item-1", "item-2", "item-3"})
	}
}

func BenchmarkLogger_child_log(b *testing.B) {
	log := newBenchmarkLogger(b)
	child := log.Child("worker", "request", "job", "benchmark")
	b.ResetTimer()

	for i := 0; b.Loop(); i++ {
		child.Infof("Request completed")
	}
}

func BenchmarkLogger_json_record(b *testing.B) {
	log := newBenchmarkLogger(b)
	record := benchmarkJSONRecord{
		UserID:  "12345",
		Name:    "Jane Doe",
		Email:   "jane.doe@example.com",
		Active:  true,
		Retries: 2,
	}
	b.ResetTimer()

	for i := 0; b.Loop(); i++ {
		log.Record("user", record).Infof("User updated")
	}
}

func BenchmarkLogger_json_redacted_record(b *testing.B) {
	log := newBenchmarkLogger(b)
	record := benchmarkJSONRecord{
		UserID:  "12345",
		Name:    "Jane Doe",
		Email:   "jane.doe@example.com",
		Active:  true,
		Retries: 2,
	}
	b.ResetTimer()

	for i := 0; b.Loop(); i++ {
		log.RecordWithKeysToRedact("user", record, "name", "email").Infof("User updated")
	}
}

func BenchmarkLogger_redacted_string(b *testing.B) {
	log := newBenchmarkLogger(b, logger.EmailRedactor)
	b.ResetTimer()

	for i := 0; b.Loop(); i++ {
		log.Infof("Authentication failed for jane.doe@example.com")
	}
}

func BenchmarkLogger_obfuscated_string(b *testing.B) {
	block, err := aes.NewCipher([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		b.Fatal(err)
	}
	log := newBenchmarkLogger(b, block)
	b.ResetTimer()

	for i := 0; b.Loop(); i++ {
		log.Infof("Authentication token: %s", log.Obfuscate("sensitive-token"))
	}
}
