# go-logger

go-logger is a logging library based on [node-bunyan](trentm/node-bunyan). 

The output is compatible with the `bunyan` log reader application from that `node` package.

## Usage

You first start by creating a `Logger` object that will operate on a `Sink`.

```go
package main

import "github.com/gildas/go-logger"

var Log = logger.Create("myapp")
```

Then, you can log messages to the same levels from `node-bunyan`:

```go
Log.Tracef("This is a message at the trace level for %s", myObject)
Log.Debugf("This is a message at the debug level for %s", myObject)
Log.Infof("This is a message at the trace level for %s", myObject)
Log.Warnf("This is a message at the warn level for %s", myObject)
Log.Errorf("This is a message at the error level for %s", myObject, err)
Log.Fatalf("This is a message at the fatal level for %s", myObject, err)
```

Note the `err` variable (must implement the [error](https://blog.golang.org/error-handling-and-go) interface) used with the last two log calls. By just adding it to the list of arguments while not mentioning it in the format string will tell the `Logger` to spit that error in a bunyan [Record field](https://github.com/trentm/node-bunyan#log-record-fields).

More generally, [Record fields](https://github.com/trentm/node-bunyan#log-record-fields) can be logged like this:

```go
Log.Record("myObject", myObject).Infof("Another message about my object")
Log.Recordf("myObject", "format %s %+v". myObject.ID(), myObject).Infof("Now the record uses a formatted value")
```

In addition to the [Bunyan core fields](https://github.com/trentm/node-bunyan#core-fields), this library adds a few Record Fields:

- `topic` can be used for stuff like types or general topics (e.g.: "http")
- `scope` can be used to scope logging within a topic, like a `func` or a portion of code.

When the `Logger` is created its `topic` and `scope` are set to "main".

Here is a simple example how [Record fields](https://github.com/trentm/node-bunyan#log-record-fields) can be used with a type:

```go
type Stuff struct {
    Field1 string
    Field2 int
    Logger *logger.Logger // So Stuff carries its own logger
}

func (s *Stuff) SetLogger(l *logger.Logger) {
    s.Logger = l.Topic("stuff").Scope("stuff")
}

func (s Stuff) DoSomething(other *OtherStuff) error {
    log := s.Logger.Scope("dosomething")

    log.Record("other", other).Infof("Need to do something")
    if err := someFunc(s, other); err != nil {
        log.Errorf("Something went wrong with other", err)
        return err
    }
    return nil
}
```

## Sinks

Sinks are like destinations for the logged data. This is where the `Logger` writes.

When creating a `Logger`, you can specify the destination it will write to:

```go
var Log = logger.CreateWithDestination("myapp", "file://path/to/myapp.log")
var Log = logger.CreateWithDestination("myapp", "stackdriver")
var Log = logger.CreateWithDestination("myapp", "gcp")
var Log = logger.CreateWithDestination("myapp", "nil")
```

The first `Logger` will write to a file, the second to Google Stackdriver, the third to Google Cloud Platform, and the fourth to nowhere (i.e. logs do not get written at all).

By default, when creating the `Logger` with:

```go
var Log = logger.Create("myapp")
```

The `Logger` will write to the standard output or the destination specified in the environment variable `LOG_DESTINATION`.

You can write your own `Sink` by implementing the `logger.Sinker` interface and create the Logger like this:

```go
var Log = logger.CreateWithSink(mySinkObject)
var Log = logger.CreateWithSink(logger.NewMultiSink(sink1, sink2, sink3))
```

If the given sink is `nil`, the `NilSink()` is used. As you may have guessed it, the second `Logger` will write simultaneously to three sinks.

The following convenience func can be used when creating a `Logger` from another one (received from arguments, for example):

```go
var Log = logger.CreateIfNil(OtherLogger, "myapp")
```

If `OtherLogger` is `nil`, the new `Logger` will write to the `NilSink()`.

## HTTP Usage

It is possible to pass `Logger` objects to [http.Handler](https://golang.org/pkg/net/http/#Handler). When doing so, the Logger will automatically write the request identifier ("X-Request-Id" HTTP Header), remote host, user agent, when the request starts and when the request finishes along with its duration.

The request identifier is attached every time the log writes in a `Record`.

Here is an example:

```go
package main

import (
    "net/http"
	"github.com/gildas/go-logger"
	"github.com/gorilla/mux"
)

func MyHandler() http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extracts the Logger from the request's context
        //  Note: Use logger.Must only when you know there is a Logger as it will panic otherwise
        log := logger.Must(logger.FromContext(r.Context()))

        log.Infof("Now we are logging inside this http Handler")
    })
}

func main() {
    log := logger.Create("myapp")
    router := mux.NewRouter()
    router.Methods("GET").Path("/").Handler(log.HttpHandler()(MyHandler()))
}
```

## To Document

Still to document:

- `func (l *Logger) Include(info)`
- `func (l *logger) GetRecord(key)`
- `func (l *logger) Child()`
- explain the `StackDriverSink` and `GCPSink`.

## TODO

- add more `Sink`, like Amazon, Azure, Elastic, etc.