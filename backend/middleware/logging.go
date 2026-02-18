package middleware

import (
	"log"
	"net/http"
	"strings"
	"time"
)

// LoggingMiddleware creates a middleware that logs requests with proxy IP support.
func LoggingMiddleware(proxyHeader string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			wrapped := &statusWriter{ResponseWriter: w, status: 200}

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)
			clientIP := getClientIP(r, proxyHeader)

			log.Printf("[%s] %s %s %d %s %s",
				clientIP,
				r.Method,
				r.URL.Path,
				wrapped.status,
				duration.Round(time.Millisecond),
				r.UserAgent(),
			)
		})
	}
}

// getClientIP extracts the client IP from the proxy header or falls back to RemoteAddr.
func getClientIP(r *http.Request, proxyHeader string) string {
	if proxyHeader != "" {
		if forwarded := r.Header.Get(proxyHeader); forwarded != "" {
			// X-Forwarded-For can contain multiple IPs, take the first one
			parts := strings.Split(forwarded, ",")
			return strings.TrimSpace(parts[0])
		}
	}

	// Fallback to RemoteAddr
	addr := r.RemoteAddr
	// Strip port if present
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}

// statusWriter wraps http.ResponseWriter to capture the status code.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
