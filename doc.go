/*
Package go-logger is a logging library based on node-bunyan, https://github.com/trentm/node-bunyan.

The output is compatible with the `bunyan` log reader application from that `node` package.

Usage


You first start by creating a `Logger` object that will operate on a `Stream`.

	package main

	import "github.com/gildas/go-logger"

	var Log = logger.Create("myapp")

Then, you can log messages to the same levels from `node-bunyan`:

	Log.Tracef("This is a message at the trace level for %s", myObject)
	Log.Debugf("This is a message at the debug level for %s", myObject)
	Log.Infof("This is a message at the trace level for %s", myObject)
	Log.Warnf("This is a message at the warn level for %s", myObject)
	Log.Errorf("This is a message at the error level for %s", myObject, err)
	Log.Fatalf("This is a message at the fatal level for %s", myObject, err)

Note the `err` variable (must implement the error interface) used with the last two log calls.
By just adding it to the list of arguments at Error or Fatal level while not mentioning it in the format string will tell
the `Logger` to spit that error in a bunyan Record field (see https://github.com/trentm/node-bunyan#log-record-fields).

More generally, Record fields can be logged like this:

	Log.Record("myObject", myObject).Infof("Another message about my object")
	Log.Recordf("myObject", "format %s %+v". myObject.ID(), myObject).Infof("This record uses a formatted value")

	log := Log.Record("dynamic", func() any { return myObject.Callme() })

	log.Infof("This is here")
	log.Infof("That is there")

In the last example, the code `myObject.Callme()` will be executed each time *log* is used to
write a message.
This is used, for example, to add a timestamp to the log's `Record`.

Here is a simple example how Record fields can be used with a type:

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

The call to Record(key, value) creates a new Logger object.
So, they are like Russian dolls when it comes down to actually
writing the log message to the output stream.

In other words, Record objects are collected from their parent's `Logger`
back to the original Logger.

For example:

	var Log   = logger.Create("test")
	var child = Log.Record("key1", "value1").Record("key2", "value2")


child will actually be something like Logger(Logger(Logger(Stream to stdout))). Though we added only 2 records.

Therefore, to optimize the number of Logger objects that are created, there are some convenience methods that can be used:

	func (s stuff) DoSomethingElse(other *OtherStuff) {
	    log := s.Logger.Child("new_topic", "new_scope", "id", other.ID(), "key1", "value1")

	    log.Infof("I am logging this stuff")

	    log.Records("key2", "value2", "key3", 12345).Warnf("Aouch that hurts!")
	}

The Child method will create one Logger that has a Record containing a topic, a scope, 2 keys (*id* and *key1*) with their values.

The Records method will create one Logger that has 2 keys (*key2* and *key3*) with their values.

For example, with these methods:

	var Log    = logger.Create("test")
	var child1 = Log.Child("topic", "scope", "key2", "value2", "key3", "value3")
	var child2 = child1.Records("key2", "value21", "key4", "value4")

*child1* will be something like Logger(Logger(Stream to stdout)). Though we added 2 records.

*child2* will be something like Logger(Logger(Logger(Stream to stdout))). Though we added 1 record to the 2 records added previously.

Stream objects

A Stream is where the Logger actually writes its Record data.

When creating a Logger, you can specify the destination it will write to:

	var Log = logger.Create("myapp", "file://path/to/myapp.log")
	var Log = logger.Create("myapp", "/path/to/myapp.log")
	var Log = logger.Create("myapp", "./localpath/to/myapp.log")
	var Log = logger.Create("myapp", "stackdriver")
	var Log = logger.Create("myapp", "gcp")
	var Log = logger.Create("myapp", "/path/to/myapp.log", "stderr")
	var Log = logger.Create("myapp", "nil")

The first three Logger objects will write to a file, the fourth to Google Stackdriver,
the fifth to Google Cloud Platform (GCP), the sixth to a file and stderr,
and the seventh to nowhere (i.e. logs do not get written at all).

By default, when creating a Logger with:

	var Log = logger.Create("myapp")

The Logger will write to the standard output or the destination specified in the environment variable LOG_DESTINATION.

You can also create a Logger by passing it a Stream object (these are equivalent to the previous code):

	var Log = logger.Create("myapp", &logger.FileStream{Path: "/path/to/myapp.log"})
	var Log = logger.Create("myapp", &logger.StackDriverStream{})
	var Log = logger.Create("myapp", &logger.NilStream{})
	var Log = logger.Create("myapp", &logger.FileStream{Path: "/path/to/myapp.log"}, &logger.StderrStream{})

A few notes:

- logger.CreateWithStream can also be used to create with one or more streams.
(Backward compatibility)

- logger.CreateWithDestination can also be used to create with one or more destinations.
(Backward compatibility)

- the StackDriverStream needs a LogID parameter or the value of the environment variable GOOGLE_PROJECT_ID.
(see Google's StackDriver documentation: https://godoc.org/cloud.google.com/go/logging#NewClient for the description of that parameter).

- NilStream is a Stream that does not write anything, all messages are lost.

- MultiStream is a Stream than can write to several streams.

- StdoutStream and FileStream are buffered by default.
Data is written from every LOG_FLUSHFREQUENCY (default 5 minutes) or when the Record's Level is at least ERROR.

- Streams convert the Record to write via a Converter. The converter is set to a default value per Stream.

You can also create a Logger with a combination of destinations and streams, AND you can even add some records right away:

	var Log = logger.Create("myapp",
	    &logger.FileStream{Path: "/path/to/myapp.log"},
	    "stackdriver",
	    NewRecord().Set("key", "value"),
	)

Setting the FilterLevel

All Stream types, except NilStream and MultiStream can use a FilterLevel. When set, Record objects that have a Level below the FilterLevel are not written to the Stream. This allows to log only stuff above *WARN* for instance.

These streams can even use a FilterLevel per topic and scope. This allows to log everything at the *INFO* level and only the log messages beloging to the topic *db* at the *DEBUG* level, for instance. Or even at the topic *db* and scope *disk*.

The FilterLevel can be set via the environment variable LOG_LEVEL:

- LOG_LEVEL=INFO

  will set the FilterLevel to *INFO*, which is the default if nothing is set;

- LOG_LEVEL=INFO;DEBUG:{topic1} or LOG_LEVEL=TRACE:{topic1};DEBUG

  will set the FilterLevel to *DEBUG* and the FilterLevel for the topic *topic1* to *TRACE* (and all the scopes under that topic);

- LOG_LEVEL=INFO;DEBUG:{topic1:scope1,scope2}

  will set the FilterLevel to *INFO* and the FilterLevel for the topic *topic1* and scopes *scope1*, *scope2* to *DEBUG* (all the other scopes under that topic will be filtered at *INFO*);

- LOG_LEVEL=INFO;DEBUG:{topic1};TRACE:{topic2}

  will set the FilterLevel to *INFO* and the FilterLevel for the topic *topic1* to *DEBUG*, respectively *topic2* and *TRACE* (and all the scopes under these topics);

- The last setting of a topic supersedes the ones set before;

- If the environment variable `DEBUG` is set to *1*, the default FilterLevel is overrident and set to *DEBUG*.

It is also possible to change the FilterLevel by calling `FilterMore()`and `FilterLess()` methods on the `Logger` or any of its `Streamer` members. The former will log less data and the latter will log more data. We provide an example of how to use these in the [examples](examples/set-level-with-signal/) folder using Unix signals.

	log := logger.Create("myapp", &logger.StdoutStream{})
	// We are filtering at INFO
	log.FilterLess()
	// We are now filtering at DEBUG

StackDriver Stream

If you plan to log to Google's StackDriver from a Google Cloud Kubernetes or a Google Cloud Instance,
you do not need the StackDriver Stream and should use the Stdout Stream with the StackDriver Converter,
since the standard output of your application will be captured automatically by Google to feed StackDriver:

	var Log = logger.Create("myapp", "gcp") // "google" or "googlecloud" are valid aliases
	var Log = logger.Create("myapp", &logger.StdoutStream{Converter: &logger.StackDriverConverter{}})


To be able to use the StackDriver Stream from outside Google Cloud, you have some configuration to do first.

On your workstation, you need to get the key filename:
1. Authenticate with Google Cloud
	gcloud auth login
2. Create a Service Account (`logger-account` is just an example of a service account name)
	gcloud iam service-acccount create logger-account
3. Associate the Service Account to the Project you want to use
	gcloud projects add-iam-policy-binding my-logging-project \
	  --member "serviceAccount:logger-account@my-logging-project.iam.gserviceaccount.com" \
	  --role "roles/logging.logWriter"
4. Retrieve the key filename
	gcloud iam service-accounts keys create /path/to/key.json \
	  --iam-account logger-account@my-logging-project.iam.gserviceaccount.com

You can either set the GOOGLE_APPLICATION_CREDENTIAL and GOOGLE_PROJECT_ID environment variables
with the path of the obtained key and Google Project ID or provide them to the StackDriver stream:

	var Log = logger.Create("myapp", &logger.StackDriverStream{})
	var Log = logger.Create("myapp", &logger.StackDriverStream{
	    Parent:      "my-logging-project",
	    KeyFilename: "/path/to/key.json",
	})

Writing your own Stream

You can also write your own Stream by implementing the logger.Streamer interface and create the Logger like this:

	var Log = logger.Create("myapp", &MyStream{})

Miscellaneous

The following convenience methods can be used when creating a Logger from another one (received from arguments, for example):

	var Log = logger.CreateIfNil(OtherLogger, "myapp")
	var Log = logger.Create("myapp", OtherLogger)

If OtherLogger is nil, the new Logger will write to the NilStream().

	var Log = logger.Must(logger.FromContext(context))

Must() can be used to create a Logger from a method that returns (*Logger, error),
if there is an error, Must will panic.

FromContext() can be used to retrieve a Logger from a Go context.
(This is used in the next paragraph about HTTP Usage).

log.ToContext will store the Logger to the given Go context.

Redacting

The Logger can redact records as needed by simply implementing the logger.Redactable interface in the data that is logged.

For example:

	type Customer {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
	}

	// implements logger.Redactable
	func (customer Customer) Redact() any {
		return Customer{customer.ID, "REDACTED"}
	}

	main() {
		// ...
		customer := Customer{uuid, "John Doe"}

		log.Record("customer", customer).Infof("Got a customer")
	}


You can also redact the log messages by providing regular expressions, called redactors. Whenever a redactor matches, its matched content is replaced with "REDACTED".

You can assign several redactors to a single logger:

	r1, err := logger.NewRedactor("[0-9]{10}")
	r2 := (logger.Redactor)(myregexp)
	log := logger.Create("test", r1, r2)

You can also add redactors to a child logger (without modifying the parent logger):

	r3 := logger.NewRedactor("[a-z]{8}")
	log := parent.Child("topic", "scope", "record1", "value1", r3)

Note: Adding redactors to a logger **WILL** have a performance impact on your application as each regular expression will be matched against every single message produced by the logger. We advise you to use as few redactors as possible and contain them in child logger, so they have a minmal impact.

Converters

The Converter object is responsible for converting the Record, given to the Stream to write, to match other log viewers.

The default Converter is BunyanConverter so the bunyan log viewer can read the logs.

Here is a list of all the converters:

- BunyanConverter, the default converter (does nothing, actually),
- CloudWatchConverter produces logs that are nicer with AWS CloudWatch log viewer.
- PinoConverter produces logs that can be used by pino (http://getpino.io),
- StackDriverConverter produces logs that are nicer with Google StackDriver log viewer,

Note: When you use converters, their output will most probably not work anymore with `bunyan`. That means you cannot have both worlds in the same Streamer. In some situation, you can survive this by using several streamers, one converted, one not.

Writing your own Converter

You can also write your own Converter by implementing the logger.Converter interface:

	type MyConverter struct {
		// ...
	}

	func (converter *MyConverter) Convert(record Record) Record {
	    record["newvalue"] = true
	    return record
	}

	var Log = logger.Create("myapp", &logger.StdoutStream{Converter: &MyConverter{}})

Standard Log Compatibility

To use a Logger with the standard go log library, you can simply call the `AsStandardLog()` method. You can optionally give a Level:

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

You can also give an io.Writer to the standard log constructor:

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

Since Writer() returns io.Writer, anything that uses that interface could, in theory, write to a Logger.

HTTP Usage

It is possible to pass Logger objects to http.Handler object. When doing so, the Logger will automatically write
the request identifier ("X-Request-Id" HTTP Header), remote host, user agent, when the request starts
and when the request finishes along with its duration.

The request identifier is attached every time the log writes in a Record.

Here is an example:

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

Environment Variables

The Logger can be configured completely by environment variables if needed. These are:

  LOG_DESTINATION, default: StdoutStream
The Stream to write logs to. It can be a comma-separated list (for a MultiStream)

  LOG_LEVEL, default: INFO
The level to filter by default. If the environment DEBUG is set the default level is DEBUG

  LOG_CONVERTER, default: bunyan
The default Converter to use

  LOG_FLUSHFREQUENCY, default: 5 minutes
The default Flush Frequency for the streams that will be buffered

  GOOGLE_APPLICATION_CREDENTIALS
The path to the credential file for the StackDriverStream

  GOOGLE_PROJECT_ID
The Google Cloud Project ID for the StackDriverStream

  DEBUG, default: none
If set to "1", this will set the default level to filter to DEBUG
*/

package logger
