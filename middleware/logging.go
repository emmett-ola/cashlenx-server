package middleware

import (
	"net/http"
	"time"

	"github.com/macar-x/cashlenx-server/util"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	return rw.ResponseWriter.Write(data)
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logging middleware to log HTTP requests
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := r.Header.Get(util.RequestIDHeader)
		if requestID == "" {
			requestID = util.NewRequestID()
		}
		w.Header().Set(util.RequestIDHeader, requestID)
		r = r.WithContext(util.ContextWithRequestID(r.Context(), requestID))

		// Wrap the response writer to capture status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Call the next handler
		next.ServeHTTP(wrapped, r)

		// Log the request using the centralized zap logger
		duration := time.Since(start)
		util.Logger.Infow(
			"HTTP Request",
			"method", r.Method,
			"path", r.RequestURI,
			"status", wrapped.statusCode,
			"duration", duration,
			"remote_addr", r.RemoteAddr,
			"request_id", requestID,
		)
	})
}
