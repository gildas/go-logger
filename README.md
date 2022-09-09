# go-logger

![GoVersion](https://img.shields.io/github/go-mod/go-version/gildas/go-logger)
[![GoDoc](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/gildas/go-logger) 
[![License](https://img.shields.io/github/license/gildas/go-logger)](https://github.com/gildas/go-logger/blob/master/LICENSE) 
[![Report](https://goreportcard.com/badge/github.com/gildas/go-logger)](https://goreportcard.com/report/github.com/gildas/go-logger)  

![master](https://img.shields.io/badge/branch-master-informational)
[![Test](https://github.com/gildas/go-logger/actions/workflows/test.yml/badge.svg?branch=master)](https://github.com/gildas/go-logger/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/gildas/go-logger/branch/master/graph/badge.svg?token=gFCzS9b7Mu)](https://codecov.io/gh/gildas/go-logger/branch/master)

![dev](https://img.shields.io/badge/branch-dev-informational)
[![Test](https://github.com/gildas/go-logger/actions/workflows/test.yml/badge.svg?branch=dev)](https://github.com/gildas/go-logger/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/gildas/go-logger/branch/dev/graph/badge.svg?token=gFCzS9b7Mu)](https://codecov.io/gh/gildas/go-logger/branch/dev)

go-logger is a logging library based on [node-bunyan](https://github.com/trentm/node-bunyan).

The output is compatible with the `bunyan` log reader application from that `node` package.

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

Note the `err` variable (must implement the [error](https://blog.golang.org/error-handling-and-go) interface) used with the last two log calls. By just adding it to the list of arguments at Error or Fatal level while not mentioning it in the format string will tell the `Logger` to spit that error in a bunyan [Record field](https://github.com/trentm/node-bunyan#log-record-fields).

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

The call to `Record(key, value)` creates a new `Logger` object. So, they are like Russian dolls when it comes down to actually writing the log message to the output stream. In other words, `Record` objects are collected from their parent's `Logger` back to the original `Logger`.  

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
var Log = logger.Create("myapp", "file://path/to/myapp.log")
var Log = logger.Create("myapp", "/path/to/myapp.log")
var Log = logger.Create("myapp", "./localpath/to/myapp.log")
var Log = logger.Create("myapp", "stackdriver")
var Log = logger.Create("myapp", "gcp")
var Log = logger.Create("myapp", "/path/to/myapp.log", "stderr")
var Log = logger.Create("myapp", "nil")
```

The first three `Logger` objects will write to a file, the fourth to Google Stackdriver, the fifth to Google Cloud Platform (GCP), the sixth to a file and stderr, and the seventh to nowhere (i.e. logs do not get written at all).

By default, when creating the `Logger` with:

```go
var Log = logger.Create("myapp")
```

The `Logger` will write to the standard output or the destination specified in the environment variable `LOG_DESTINATION`.

You can also create a `Logger` by passing it a `Stream` object (these are equivalent to the previous code):

```go
var Log = logger.Create("myapp", &logger.FileStream{Path: "/path/to/myapp.log"})
var Log = logger.Create("myapp", &logger.StackDriverStream{})
var Log = logger.Create("myapp", &logger.NilStream{})
var Log = logger.Create("myapp", &logger.FileStream{Path: "/path/to/myapp.log"}, &logger.StderrStream{})
```

A few notes:
- `logger.CreateWithStream` can also be used to create with one or more streams.  
  (Backward compatibility)
- `logger.CreateWithDestination` can also be used to create with one or more destinations.  
  (Backward compatibility)
- the `StackDriverStream` needs a `LogID` parameter or the value of the environment variable `GOOGLE_PROJECT_ID`. (see [Google's StackDriver documentation](https://godoc.org/cloud.google.com/go/logging#NewClient) for the description of that parameter).
- `NilStream` is a `Stream` that does not write anything, all messages are lost.
- `MultiStream` is a `Stream` than can write to several streams.
- `StdoutStream` and `FileStream` are buffered by default. Data is written from every `LOG_FLUSHFREQUENCY` (default 5 minutes) or when the `Record`'s `Level` is at least *ERROR*.
- Streams convert the `Record` to write via a `Converter`. The converter is set to a default value per Stream.

You can also create a `Logger` with a combination of destinations and streams, AND you can even add some records right away:

```go
var Log = logger.Create("myapp",
    &logger.FileStream{Path: "/path/to/myapp.log"},
    "stackdriver",
    NewRecord().Set("key", "value"),
)
```
### Setting the FilterLevel

All `Stream` types, except `NilStream` and `MultiStream` can use a `FilterLevel`. When set, `Record` objects that have a `Level` below the `FilterLevel` are not written to the `Stream`. This allows to log only stuff above *WARN* for instance.

These streams can even use a `FilterLevel` per `topic` and `scope`. This allows to log everything at the *INFO* level and only the log messages beloging to the topic *db* at the *DEBUG* level, for instance. Or even at the topic *db* and scope *disk*.

The `FilterLevel` can be set via the environment variable `LOG_LEVEL`:

- `LOG_LEVEL=INFO`  
  will set the FilterLevel to *INFO*, which is the default if nothing is set;
- `LOG_LEVEL=INFO;DEBUG:{topic1}` or `LOG_LEVEL=TRACE:{topic1};DEBUG`  
  will set the FilterLevel to *DEBUG* and the FilterLevel for the topic *topic1* to *TRACE* (and all the scopes under that topic);
- `LOG_LEVEL=INFO;DEBUG:{topic1:scope1,scope2}`  
  will set the FilterLevel to *INFO* and the FilterLevel for the topic *topic1* and scopes *scope1*, *scope2* to *DEBUG* (all the other scopes under that topic will be filtered at *INFO*);
- `LOG_LEVEL=INFO;DEBUG:{topic1};TRACE:{topic2}`  
  will set the FilterLevel to *INFO* and the FilterLevel for the topic *topic1* to *DEBUG*, respectively *topic2* and *TRACE* (and all the scopes under these topics);
- The last setting of a topic supersedes the ones set before;
- If the environment variable `DEBUG` is set to *1*, the default FilterLevel is overrident and set to *DEBUG*.

It is also possible to change the FilterLevel by calling `FilterMore()`and `FilterLess()` methods on the `Logger` or any of its `Streamer` members. The former will log less data and the latter will log more data. We provide an example of how to use these in the [examples](examples/set-level-with-signal/) folder using Unix signals.

```go
log := logger.Create("myapp", &logger.StdoutStream{})
// We are filtering at INFO
log.FilterLess()
// We are now filtering at DEBUG
```

### StackDriver Stream

If you plan to log to Google's StackDriver from a Google Cloud Kubernetes or a Google Cloud Instance, you do not need the StackDriver Stream and should use the Stdout Stream with the StackDriver Converter, since the standard output of your application will be captured automatically by Google to feed StackDriver:  
```go
var Log = logger.Create("myapp", "gcp") // "google" or "googlecloud" are valid aliases
var Log = logger.Create("myapp", &logger.StdoutStream{Converter: &logger.StackDriverConverter{}})
```

To be able to use the StackDriver Stream from outside Google Cloud, you have some configuration to do first.

On your workstation, you need to get the key filename:  
1. Authenticate with Google Cloud  
```console
gcloud auth login
```
2. Create a Service Account (`logger-account` is just an example of a service account name)  
```console
gcloud iam service-acccount create logger-account
```
3. Associate the Service Account to the Project you want to use  
```console
gcloud projects add-iam-policy-binding my-logging-project \
  --member "serviceAccount:logger-account@my-logging-project.iam.gserviceaccount.com" \
  --role "roles/logging.logWriter"
```
4. Retrieve the key filename  
```console
gcloud iam service-accounts keys create /path/to/key.json \
  --iam-account logger-account@my-logging-project.iam.gserviceaccount.com
```

You can either set the `GOOGLE_APPLICATION_CREDENTIAL` and `GOOGLE_PROJECT_ID` environment variables with the path of the obtained key and Google Project ID or provide them to the StackDriver stream:  
```go
var log = logger.Create("myapp", &logger.StackDriverStream{})
```

```go
var log = logger.Create("myapp", &logger.StackDriverStream{
    Parent:      "my-logging-project",
    KeyFilename: "/path/to/key.json",
})
```

### Writing your own Stream

You can also write your own `Stream` by implementing the `logger.Streamer` interface and create the Logger like this:

```go
var log = logger.Create("myapp", &MyStream{})
```

### Logging Source Information

It is possible to log source information such as the source filename and code line, go package, and the caller func.

```go
var Log = logger.Create("myapp", &logger.FileStream{Path: "/path/to/myapp.log", SourceInfo: true})

func MyFunc() {
  Log.Infof("I am Here")
}
```

**Note**: Since this feature can be expensive to compute, it is turned of by default.  
To turn it on, you need to either specify the option in the Stream object, set the environment variable `LOG_SOURCEINFO` to _true_. It is also turned on if the environment variable `DEBUG` is _true_.

### Timing your funcs

You can automatically log the duration of your func by calling them via the logger:

```go
log.TimeFunc("message shown with the duration", func() {
  log.Info("I am here")
  // ... some stuff that takes time
  time.Sleep(12*time.Second)
})
```

The duration will logged in the `msg` record after the given message. It will also be added as a float value in the `duration` record.

There are 3 more variations for funcs that return an error, a value, an error and a value:

```go
result := log.TimeFuncV("message shown with the duration", func() interface{} {
  log.Info("I am here")
  // ... some stuff that takes time
  time.Sleep(12*time.Second)
  return 12
})

err := log.TimeFuncE("message shown with the duration", func() err {
  log.Info("I am here")
  // ... some stuff that takes time
  time.Sleep(12*time.Second)
  return errors.ArgumentMissing.With("path")
})

result, err := log.TimeFuncV("message shown with the duration", func() (interface{}, error) {
  log.Info("I am here")
  // ... some stuff that takes time
  time.Sleep(12*time.Second)
  return 12, errors.ArgumentInvalid.With("value", 12)
})
```

### Miscellaneous

The following convenience methods can be used when creating a `Logger` from another one (received from arguments, for example):

```go
var log = logger.CreateIfNil(OtherLogger, "myapp")
```
```go
var log = logger.Create("myapp", OtherLogger)
```

If `OtherLogger` is `nil`, the new `Logger` will write to the `NilStream()`.

```go
var log = logger.Must(logger.FromContext(context))
```

`Must` can be used to create a `Logger` from a method that returns `*Logger, error`, if there is an error, `Must` will panic.

`FromContext` can be used to retrieve a `Logger` from a GO context. (This is used in the paragraph about HTTP Usage)  

`log.ToContext` will store the `Logger` to the given GO context.

## Redacting

The `Logger` can redact records as needed by simply implementing the `logger.Redactable` interface in the data that is logged.

For example:
```go
type Customer {
  ID   uuid.UUID `json:"id"`
  Name string    `json:"name"`
}

// implements logger.Redactable
func (customer Customer) Redact() interface{} {
  return Customer{customer.Name, "REDACTED"}
}

main() {
  // ...
  customer := Customer{uuid, "John Doe"}

  log.Record("customer", customer).Infof("Got a customer")
}
```

You can also redact the log messages by providing regular expressions, called redactors. Whenever a redactor matches, its matched content is replaced with "REDACTED".

You can assign several redactors to a single logger:

```go
r1, err := logger.NewRedactor("[0-9]{10}")
r2 := (logger.Redactor)(myregexp)
log := logger.Create("test", r1, r2)
```

You can also add redactors to a child logger (without modifying the parent logger):

```go
r3 := logger.NewRedactor("[a-z]{8}")
log := parent.Child("topic", "scope", "record1", "value1", r3)
```

**Note:** Adding redactors to a logger **WILL** have a performance impact on your application as each regular expression will be matched against every single message produced by the logger. We advise you to use as few redactors as possible and contain them in child logger, so they have a minimal impact.

## Converters

The `Converter` object is responsible for converting the `Record`, given to the `Stream` to write, to match other log viewers.

The default `Converter` is `BunyanConverter` so the `bunyan` log viewer can read the logs.

Here is a list of all the converters:

- `BunyanConverter`, the default converter (does nothing, actually),
- `CloudWatchConverter` produces logs that are nicer with AWS CloudWatch log viewer.
- `PinoConverter` produces logs that can be used by [pino](http://getpino.io),
- `StackDriverConverter` produces logs that are nicer with Google StackDriver log viewer,

**Note**: When you use converters, their output will most probably not work anymore with `bunyan`. That means you cannot have both worlds in the same Streamer. In some situation, you can survive this by using several streamers, one converted, one not.

### Writing your own Converter

You can also write your own `Converter` by implementing the `logger.Converter` interface:  

```go
type MyConverter struct {
	// ...
}

func (converter *MyConverter) Convert(record Record) Record {
    record["newvalue"] = true
    return record
}

var Log = logger.Create("myapp", &logger.StdoutStream{Converter: &MyConverter{}})
```

## Standard Log Compatibility

To use a `Logger` with the standard go `log` library, you can simply call the `AsStandardLog()` method. You can optionally give a `Level`:  
```go
package main

import (
  "net/http"
	"github.com/gildas/go-logger"
)

func main() {
    log := logger.Create("myapp")

    server1 := http.Server{
      // extra http stuff
      ErrorLog: log.AsStandardLog()
    }

    server2 := http.Server{
      // extra http stuff
      ErrorLog: log.AsStandardLog(logger.WARN)
    }
}
```

You can also give an `io.Writer` to the standard `log` constructor:  
```go
package main

import (
  "log"
  "net/http"
	"github.com/gildas/go-logger"
)

func main() {
    mylog := logger.Create("myapp")

    server1 := http.Server{
      // extra http stuff
      ErrorLog: log.New(mylog.Writer(), "", 0),
    }

    server2 := http.Server{
      // extra http stuff
      ErrorLog: log.New(mylog.Writer(logger.WARN), "", 0),
    }
}
```

Since `Writer()` returns `io.Writer`, anything that uses that interface could, in theory, write to a `Logger`.

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

When the http request handler (_MyHandler_) starts, the following records are logged:  
- `reqid`, contains the request Header X-Request-Id if present, or a random UUID
- `path`, contains the URL Path of the request
- `remote`, contains the remote address of the request
- The `topic` is set to "route" and the `scope` to the path of the request URL

When the http request handler (_MyHandler_) ends, the following additional records are logged:  
- `duration`, contains the duration in seconds (**float64**) of the handler execution

## Environment Variables

The `Logger` can be configured completely by environment variables if needed. These are:  
- `LOG_DESTINATION`, default: `StdoutStream`  
  The `Stream`s to write logs to. It can be a comma-separated list (for a `MultiStream`)
- `LOG_LEVEL`, default: *INFO*  
  The level to filter by default. If the environment `DEBUG` is set the default level is *DEBUG*
- `LOG_CONVERTER`, default: "bunyan"  
  The default `Converter` to use
- `LOG_FLUSHFREQUENCY`, default: 5 minutes  
  The default Flush Frequency for the streams that will be buffered
- `LOG_OBFUSCATION_KEY`, default: none  
  The SSL public key to use when obfuscating if you want a reversible obfuscation
- `GOOGLE_APPLICATION_CREDENTIALS`  
  The path to the credential file for the `StackDriverStream`
- `GOOGLE_PROJECT_ID`  
  The Google Cloud Project ID for the `StackDriverStream`
- `DEBUG`, default: none  
  If set to "1", this will set the default level to filter to *DEBUG*

# Thanks

Special thanks to [@chakrit](https://github.com/chakrit) for his [chakrit/go-bunyan](https://github.com/chakrit/go-bunyan) that inspired me. In fact earlier versions were wrappers around his library.  

Well, we would not be anywhere without the original work of [@trentm](https://github.com/trentm) and the original [trentm/node-bunyan](https://github.com/trentm/node-bunyan). Many, many thanks!
