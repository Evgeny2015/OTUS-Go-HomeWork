package internalhttp

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func loggingMiddleware(logger Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		// Format log line according to specification
		ip := extractIP(r.RemoteAddr)
		date := start.Format("[02/Jan/2006:15:04:05 -0700]")
		method := r.Method
		path := r.URL.RequestURI()
		proto := r.Proto
		status := strconv.Itoa(rw.statusCode)
		latency := strconv.FormatInt(duration.Milliseconds(), 10)
		userAgent := r.UserAgent()
		if userAgent == "" {
			userAgent = "-"
		} else {
			userAgent = `"` + userAgent + `"`
		}

		logLine := strings.Join([]string{
			ip,
			date,
			method,
			path,
			proto,
			status,
			latency,
			userAgent,
		}, " ")

		logger.Info(logLine)
	})
}

// extractIP extracts the IP address from RemoteAddr (ip:port).
func extractIP(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		// If splitting fails, assume it's already an IP
		return addr
	}
	return host
}

// responseWriter wraps http.ResponseWriter to capture status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
