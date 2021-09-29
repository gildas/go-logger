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
			reqLogger := l.Record("topic", "route").Record("scope", r.URL.Path).Record("reqid", reqid)
			reqLogger.
				Record("remote", r.RemoteAddr).
				Record("UserAgent", r.UserAgent()).
				Infof("request start: %s %s", r.Method, html.EscapeString(r.URL.Path))

			// Adding reqid and reqLogger to r.Context and serving the request
			//nolint:staticcheck
			next.ServeHTTP(w, r.WithContext(reqLogger.ToContext(context.WithValue(r.Context(), "reqid", reqid))))

			// Logging the duration of the request handling
			reqLogger.
				Record("duration", time.Since(start).Seconds()).
				Infof("request finish: %s %s", r.Method, html.EscapeString(r.URL.Path))
		})
	}
}
