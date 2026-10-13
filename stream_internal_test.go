package logger

import (
	"context"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"cloud.google.com/go/logging/apiv2/loggingpb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	googleoption "google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestCreateUnbufferedStreamWithDEBUG(t *testing.T) {
	currentDEBUG := os.Getenv("DEBUG")
	defer func() { _ = os.Setenv("DEBUG", currentDEBUG) }()

	_ = os.Setenv("DEBUG", "1")
	stream := CreateStream(NewLevelSet(DEBUG))
	assert.True(t, stream.(*StdoutStream).Unbuffered, "Stdout should be unbuffered with DEBUG set")
}

func TestMultiStreamCanSetFilterLevel(t *testing.T) {
	stream := &MultiStream{streams: []Streamer{&StdoutStream{}, &StderrStream{}}}
	assert.Equal(t, UNSET, stream.streams[0].(*StdoutStream).FilterLevels.GetDefault())
	assert.Equal(t, UNSET, stream.streams[1].(*StderrStream).FilterLevels.GetDefault())
	stream.SetFilterLevel(WARN)
	assert.Equal(t, WARN, stream.streams[0].(*StdoutStream).FilterLevels.GetDefault())
	assert.Equal(t, WARN, stream.streams[1].(*StderrStream).FilterLevels.GetDefault())
}

type fakeStackDriverServer struct {
	loggingpb.UnimplementedLoggingServiceV2Server
	mutex   sync.Mutex
	entries []*loggingpb.LogEntry
}

func (server *fakeStackDriverServer) WriteLogEntries(ctx context.Context, request *loggingpb.WriteLogEntriesRequest) (*loggingpb.WriteLogEntriesResponse, error) {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for _, entry := range request.Entries {
		if len(entry.LogName) == 0 {
			entry.LogName = request.LogName
		}
		server.entries = append(server.entries, entry)
	}
	return &loggingpb.WriteLogEntriesResponse{}, nil
}

// Entries gets the entries sent to the given log (ignoring the client's diagnostic entries)
func (server *fakeStackDriverServer) Entries(logName string) []*loggingpb.LogEntry {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	entries := []*loggingpb.LogEntry{}
	for _, entry := range server.entries {
		if entry.LogName == logName {
			entries = append(entries, entry)
		}
	}
	return entries
}

func startFakeStackDriverServer(t *testing.T) (*fakeStackDriverServer, string) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err, "Failed to listen")
	fake := &fakeStackDriverServer{}
	server := grpc.NewServer()
	loggingpb.RegisterLoggingServiceV2Server(server, fake)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)
	return fake, listener.Addr().String()
}

func TestCanStreamToFakeStackDriver(t *testing.T) {
	t.Setenv("LOG_LEVEL", "WARN")
	t.Setenv("GCE_METADATA_HOST", "127.0.0.1:1") // Avoid waiting for the GCE metadata server detection
	fake, address := startFakeStackDriverServer(t)
	stream := &StackDriverStream{
		LogID:  "test",
		Parent: "projects/go-logger-test",
		clientOptions: []googleoption.ClientOption{
			googleoption.WithEndpoint(address),
			googleoption.WithoutAuthentication(),
			googleoption.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
		},
	}
	err := stream.Write(NewRecord().Set("bello", "banana").Set("level", WARN).Set("time", time.Now()).Set("msg", "Hello 01"))
	require.NoError(t, err, "Failed to write to stream")
	assert.Equal(t, WARN, stream.FilterLevels.GetDefault(), "Filter levels should have been read from the environment")
	stream.Flush()
	entries := fake.Entries("projects/go-logger-test/logs/test")
	require.Len(t, entries, 1, "The fake server should have received 1 entry")
	assert.Equal(t, "Hello 01", entries[0].GetJsonPayload().GetFields()["message"].GetStringValue())
	assert.Equal(t, "banana", entries[0].GetJsonPayload().GetFields()["bello"].GetStringValue())

	err = stream.Write(NewRecord().Set("bello", "mata banana").Set("level", ERROR).Set("time", time.Now()).Set("msg", "Hello 02"))
	require.NoError(t, err, "Failed to write to stream")
	stream.Close()
	entries = fake.Entries("projects/go-logger-test/logs/test")
	require.Len(t, entries, 2, "The fake server should have received 2 entries")
	assert.Equal(t, "Hello 02", entries[1].GetJsonPayload().GetFields()["message"].GetStringValue())
}
