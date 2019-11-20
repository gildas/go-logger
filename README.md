# go-logger

go-logger is a logging library based on [node-bunyan](trentm/node-bunyan). 

The output is compatible with the `bunyan` log reader application from that `node` package.

[![Build Status](https://dev.azure.com/keltiek/gildas/_apis/build/status/gildas.go-logger?branchName=master)](https://dev.azure.com/keltiek/gildas/_build/latest?definitionId=1&branchName=master)

## Usage

You first start by creating a `Logger` object that will operate on a `Stream`.

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
Log.Recordf("myObject", "format %s %+v". myObject.ID(), myObject).Infof("This record uses a formatted value")

log := Log.Record("dynamic", func() interface{} { return myObject.Callme() })

log.Infof("This is here")
log.Infof("That is there")
```

In the last example, the code `myObject.Callme()` will be executed each time *log* is used to write a message.
This is used, as an example, to add a timestamp to the log's `Record`.

In addition to the [Bunyan core fields](https://github.com/trentm/node-bunyan#core-fields), this library adds a couple of Record Fields:

- `topic` can be used for stuff like types or general topics (e.g.: "http")
- `scope` can be used to scope logging within a topic, like a `func` or a portion of code.

When the `Logger` is created its *topic* and *scope* are set to "main".

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

The call to `Record(key, value)` creates a new `Logger` object. So, they are like Russian dolls when it comes down to actually writing the log message to the output stream. In other words, `Record` are collected from their parent's `Logger` back to the original `Logger`.  

For example:  
```go
var Log   = logger.Create("test")
var child = Log.Record("key1", "value1").Record("key2", "value2")
```

*child* will actually be something like `Logger(Logger(Logger(Stream to stdout)))`. Though we added only 2 records.  

Therefore, to optimize the number of `Logger` objects that are created, there are some convenience methods that can be used:

```go
func (s stuff) DoSomethingElse(other *OtherStuff) {
    log := s.Logger.Child("new_topic", "new_scope", "id", other.ID(), "key1", "value1")

    log.Infof("I am logging this stuff")

    log.Records("key2", "value2", "key3", 12345).Warnf("Aouch that hurts!")
}
```

The `Child` method will create one `Logger` that has a `Record` containing a topic, a scope, 2 keys (*id* and *key1*) with their values.

The `Records` method will create one `Logger` that has 2 keys (*key2* and *key3*) with their values.

For example, with these methods:  
```go
var Log    = logger.Create("test")
var child1 = Log.Child("topic", "scope", "key2", "value2", "key3", "value3")
var child2 = child1.Records("key2", "value21", "key4", "value4")
```

*child1* will be something like `Logger(Logger(Stream to stdout))`. Though we added 2 records.  
*child2* will be something like `Logger(Logger(Logger(Stream to stdout)))`. Though we added 1 record to the 2 records added previously.  

## Stream objects

A `Stream` is where the `Logger` actually writes its `Record` data.

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

You can also create a `Logger` by passing it a `Stream` object (these are equivalent to the previous code):

```go
var Log = logger.CreateWithStream("myapp", &logger.FileStream{Path: "/path/to/myapp.log"})
var Log = logger.CreateWithStream("myapp", &logger.StackDriverStream{})
var Log = logger.CreateWithStream("myapp", &logger.GCPStream{})
var Log = logger.CreateWithStream("myapp", &logger.NilStream{})
```

A few notes:
- the `StackDriverStream` needs a `ProjectID` parameter or the value of the environment variable `PROJECT_ID`.  
  It can use a `LogID` (see Google's StackDriver documentation).
- `NilStream` is a `Stream` that does not write anything, all messages are lost.
- `MultiStream` is a `Stream` than can write to several streams.
- All `Stream` types, except `NilStream` and `MultiStream` can use a `FilterLevel`. When set, `Record` objects that have a `Level` below the `FilterLevel` are not written to the `Stream`. This allows to log only stuff above *Warn* for instance. The `FilterLevel` can be set via the environment variable `LOG_LEVEL`.
- `StdoutStream` and `FileStream` are buffered by default. Data is written from every `LOG_FLUSHFREQUENCY` (default 5 minutes) or when the `Record`'s `Level` is at least *ERROR*.

You can write your own `Stream` by implementing the `logger.Streamer` interface and create the Logger like this:

```go
var Log = logger.CreateWithStream("myapp", &MyStream{})
```

The following convenience methods can be used when creating a `Logger` from another one (received from arguments, for example):

```go
var Log = logger.CreateIfNil(OtherLogger, "myapp")
```

If `OtherLogger` is `nil`, the new `Logger` will write to the `NilStream()`.

```go
var Log = logger.Must(logger.FromContext(context))
```

`Must` can be used to create a `Logger` from a method that returns `*Logger, error`, if there is an error, `Must`will `panic`.

`FromContext` can be used to retrieve a `Logger` from a GO context. (This is used in the next paragraph)  
`log.ToContext` will store the `Logger` to the given GO context.

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

# Thanks

Special thanks to [@chakrit](https://github.com/chakrit) for his [chakrit/go-bunyan](https://github.com/chakrit/go-bunyan) that inspired me. In fact earlier versions were wrappers around his library.  

Well, we would not be anywhere without the original work of [@trentm](https://github.com/trentm) and the original [trentm/node-bunyan](https://github.com/trentm/node-bunyan). Many, many thanks!
