package middlewares

import (
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/skuid/spec/version"
)

func monitor(verb, path string, httpCode int, reqStart time.Time) {
	elapsed := float64((time.Since(reqStart)) / time.Microsecond)

	statsdClient := Client()

	// datadog statsd
	if statsdClient != nil {
		tags := [4]string{
			fmt.Sprintf("%s:%s", "sha", version.Commit),
			fmt.Sprintf("%s:%s", "method", strings.ToLower(verb)),
			fmt.Sprintf("%s:%s", "path", path),
			fmt.Sprintf("%s:%d", "status", httpCode),
		}
		statsdClient.Incr("http_request_count", tags[:], 1)
		statsdClient.Histogram("http_request_duration", elapsed, tags[:3], 1)

		statsdClient.Incr(fmt.Sprintf("http_request_status_%s", statusType(httpCode)), tags[:], 1)
	}
}

func statusType(code int) string {
	switch math.Floor(float64(code) / float64(100)) {
	case 5:
		return "server_error"
	case 4:
		return "client_error"
	case 3:
		return "redirection"
	case 2:
		return "successful"
	case 1:
		return "informational"
	default:
		return "unknown_error"
	}
}

// InstrumentRoute is a middleware for adding metrics to a route.
// The following metrics are added:
//	# Counter
//	http_request_count{"verb", "path"}
//	# Counter
//	http_request_status_%s{"verb", "path"} // where %s is each specific HTTP status code
//	# Histogram
//	http_request_duration{"verb", "path"}
func InstrumentRoute() Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			now := time.Now()
			wrappedWriter := &statusLoggingResponseWriter{w, http.StatusOK, 0}

			defer func() {
				monitor(r.Method, r.URL.Path, wrappedWriter.status, now)
			}()
			h.ServeHTTP(wrappedWriter, r)
		})
	}
}
