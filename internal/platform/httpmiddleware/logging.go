package httpmiddleware

import (
	"log/slog"
	"net/http"
	"time"
)

// responseRecorder forwards the response to the client and remembers its final status.
// A zero statusCode means no final status has been written yet.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader sends the status and remembers the first final response code.
// Informational responses are forwarded without setting the final status.
func (r *responseRecorder) WriteHeader(statusCode int) {
	if r.statusCode != 0 {
		return
	}
	if statusCode >= 100 && statusCode <= 199 {
		r.ResponseWriter.WriteHeader(statusCode)
		return
	}
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

// Write sends the response body, recording 200 if no final status was set beforehand.
// It returns the number of bytes written and any error from the underlying writer.
func (r *responseRecorder) Write(body []byte) (int, error) {
	if r.statusCode == 0 {
		r.WriteHeader(http.StatusOK)
	}
	return r.ResponseWriter.Write(body)
}

// Logging records the request method, path, response status, and elapsed time after next returns.
// It leaves query parameters, headers, and request bodies out of the log.
func Logging(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &responseRecorder{
			ResponseWriter: w,
		}
		next.ServeHTTP(rec, r)

		status := rec.statusCode
		// A handler that returns without writing a response gets the default status of 200.
		if status == 0 {
			status = http.StatusOK
		}

		logger.InfoContext(
			r.Context(),
			"HTTP request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", status,
			"duration", time.Since(start),
		)
	})
}
