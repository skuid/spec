package middlewares

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/skuid/spec/version"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func monitor(verb, path string, httpCode int, reqStart time.Time) {
	elapsed := float64((time.Since(reqStart)) / time.Microsecond)

	tags := [4]string{
		fmt.Sprintf("%s:%s", "sha", version.Commit),
		fmt.Sprintf("%s:%s", "method", strings.ToLower(verb)),
		fmt.Sprintf("%s:%s", "path", path),
		fmt.Sprintf("%s:%d", "status", httpCode),
	}

	// datadog statsd
	statsdClient := Client()
	if statsdClient != nil {
		statsdClient.Incr("http_request_count", tags[:], 1)
		statsdClient.Histogram("http_request_duration", elapsed, tags[:3], 1)

		statsdClient.Incr(fmt.Sprintf("http_request_status_%s", statusType(httpCode)), tags[:], 1)
	}

	// opentelemetry metrics
	// If global meter provider is not set, GetMeterProvider returns a no-op provider.
	oMeter := otel.GetMeterProvider().Meter(
		"http_metrics",
		metric.WithInstrumentationAttributes(
			attribute.StringSlice("tags", tags[:]),
		),
	)
	oCount, err := oMeter.Int64Counter("http_request_count")
	if err != nil {
		return
	}
	oDuration, err := oMeter.Float64Histogram("http_request_duration")
	if err != nil {
		return
	}
	oStatusCount, err := oMeter.Int64Counter(fmt.Sprintf("http_request_status_%s", statusType(httpCode)))
	if err != nil {
		return
	}
	ctx := context.Background()
	oCount.Add(ctx, 1)
	oDuration.Record(ctx, elapsed)
	oStatusCount.Add(ctx, 1)
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
//
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
