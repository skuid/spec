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

type counterType int64

const (
	intCounter counterType = iota
	floatHistogram
)

const MetricsName = "http_metrics"

var meters = map[string]metric.Meter{} // key: meterName
var counters = map[string]any{}        // key: meterName.counterName

// HttpMeter will record metrics according to the ctype.
// It stores meters and counters etc. in the above global maps for re-use.
func HttpMeter(meterName, counterName string, ctype counterType, ival any, attr ...attribute.KeyValue) (err error) {
	var meter metric.Meter
	var ok bool
	ctx := context.Background()
	if meter, ok = meters[meterName]; !ok {
		meter = otel.Meter(meterName, metric.WithInstrumentationAttributes(attr...))
		meters[meterName] = meter
	}

	var counterI any
	counterKey := fmt.Sprintf("%s.%s", meterName, counterName)
	counterI, ok = counters[counterKey]

	switch ctype {
	case intCounter:
		var counter metric.Int64Counter
		if !ok {
			counter, err = meter.Int64Counter(counterName)
			if err != nil {
				return
			}
			counters[counterKey] = counter
		} else {
			var iok bool
			counter, iok = counterI.(metric.Int64Counter)
			if !iok {
				return fmt.Errorf("counter %T is not an int64 counter", counterI)
			}
		}
		var val int
		val, ok = ival.(int)
		if !ok {
			return fmt.Errorf("ival %T is not an int value", ival)
		}
		counter.Add(ctx, int64(val))
	case floatHistogram:
		var counter metric.Float64Histogram
		if !ok {
			counter, err = meter.Float64Histogram(counterName)
			if err != nil {
				return
			}
			counters[counterKey] = counter
		} else {
			var iok bool
			counter, iok = counterI.(metric.Float64Histogram)
			if !iok {
				return fmt.Errorf("counter %T is not a float64 histogram", counterI)
			}
		}
		var val float64
		val, ok = ival.(float64)
		if !ok {
			return fmt.Errorf("ival %T is not a float64 value", ival)
		}
		counter.Record(ctx, val)
	}

	return nil
}

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
	attr := attribute.StringSlice("tags", tags[:])
	var err error
	err = HttpMeter(MetricsName, "http_request_count", intCounter, 1, attr)
	if err != nil {
		panic(err)
	}
	err = HttpMeter(MetricsName, "http_request_duration", floatHistogram, elapsed, attr)
	if err != nil {
		panic(err)
	}
	err = HttpMeter(MetricsName, fmt.Sprintf("http_request_status_%s", statusType(httpCode)), intCounter, 1, attr)
	if err != nil {
		panic(err)
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
