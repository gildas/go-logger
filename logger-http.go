package logger

import (
	"bufio"
	"context"
	"html"
	"net"
	"net/http"
	"time"

	"github.com/gildas/go-errors"
	"github.com/google/uuid"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    uint64
}

// WriteHeader sends an HTTP response header with the provided
// status code.
//
// If WriteHeader is not called explicitly, the first call to Write
// will trigger an implicit WriteHeader(http.StatusOK).
// Thus explicit calls to WriteHeader are mainly used to
// send error codes or 1xx informational responses.
//
// The provided code must be a valid HTTP 1xx-5xx status code.
// Any number of 1xx headers may be written, followed by at most
// one 2xx-5xx header. 1xx headers are sent immediately, but 2xx-5xx
// headers may be buffered. Use the Flusher interface to send
// buffered data. The header map is cleared when 2xx-5xx headers are
// sent, but not with 1xx headers.
//
// The server will automatically send a 100 (Continue) header
// on the first read from the request body if the request has
// an "Expect: 100-continue" header.
func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// Hijack implements the http.Hijacker interface
//
// Hijack lets the caller take over the connection.
//
// This is used by websockets (among others)
func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, errors.Join(errors.New("ResponseWrite does not implement http.Hijaker"), errors.InvalidType.With("responseWriter", "Hijacker"))
}

/*
// Not sure yet if we need this

func (w *responseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, errors.NotImplemented.WithStack()
}

func (w *responseWriter) RoundTrip(r *http.Request) (*http.Response, error) {
	if roundtripper, ok := w.ResponseWriter.(http.RoundTripper); ok {
		return roundtripper.RoundTrip(r)
	}
	return nil, errors.NotImplemented.WithStack()
}
*/

// Write writes the data to the connection as part of an HTTP reply.
//
// If WriteHeader has not yet been called, Write calls
// WriteHeader(http.StatusOK) before writing the data. If the Header
// does not contain a Content-Type line, Write adds a Content-Type set
// to the result of passing the initial 512 bytes of written data to
// DetectContentType. Additionally, if the total size of all written
// data is under a few KB and there are no Flush calls, the
// Content-Length header is added automatically.
//
// Depending on the HTTP protocol version and the client, calling
// Write or WriteHeader may prevent future reads on the
// Request.Body. For HTTP/1.x requests, handlers should read any
// needed request body data before writing the response. Once the
// headers have been flushed (due to either an explicit Flusher.Flush
// call or writing enough data to trigger a flush), the request body
// may be unavailable. For HTTP/2 requests, the Go HTTP server permits
// handlers to continue to read the request body while concurrently
// writing the response. However, such behavior may not be supported
// by all HTTP/2 clients. Handlers should read before writing if
// possible to maximize compatibility.
func (w *responseWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.written += uint64(n)
	return n, err
}

// HttpHandler function will wrap an http handler with extra logging information
func (l *Logger) HttpHandler() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Get a request identifier and pass it to the response writer
			reqid := r.Header.Get("X-Line-Request-Id")
			if len(reqid) == 0 {
				reqid = r.Header.Get("X-Request-Id")
			}
			if len(reqid) == 0 {
				reqid = uuid.Must(uuid.NewRandom()).String()
			}
			w.Header().Set("X-Request-Id", reqid)

			// Get a new Child logger tailored to the request
			reqLogger := l.Child("route", r.URL.Path, "reqid", reqid, "path", r.URL.Path, "remote", r.RemoteAddr)
			reqLogger.
				Record("agent", r.UserAgent()).
				Record("verb", r.Method).
				Infof("request start: %s %s", r.Method, html.EscapeString(r.URL.Path))

			// Wrap the response writer to capture the status code
			writer := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Adding reqid and reqLogger to r.Context and serving the request
			//nolint:staticcheck
			next.ServeHTTP(writer, r.WithContext(reqLogger.ToContext(context.WithValue(r.Context(), "reqid", reqid))))

			// Logging the duration of the request handling
			duration := time.Since(start)
			if writer.statusCode >= 400 {
				reqLogger.
					Record("duration", duration.Seconds()).
					Record("http_status", writer.statusCode).
					Record("written", writer.written).
					Errorf("request finish: %s %s in %s", r.Method, html.EscapeString(r.URL.Path), duration)
			} else {
				reqLogger.
					Record("duration", duration.Seconds()).
					Record("http_status", writer.statusCode).
					Record("written", writer.written).
					Infof("request finish: %s %s in %s", r.Method, html.EscapeString(r.URL.Path), duration)
			}
		})
	}
}
