package logger

import (
	"context"
	"html"
	"net/http"
	"time"

	"github.com/google/uuid"
)

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

			// Adding reqid and reqLogger to r.Context and serving the request
			//nolint:staticcheck
			next.ServeHTTP(w, r.WithContext(reqLogger.ToContext(context.WithValue(r.Context(), "reqid", reqid))))

			// Logging the duration of the request handling
			duration := time.Since(start)
			reqLogger.
				Record("duration", duration.Seconds()).
				Infof("request finish: %s %s in %s", r.Method, html.EscapeString(r.URL.Path), duration)
		})
	}
}
