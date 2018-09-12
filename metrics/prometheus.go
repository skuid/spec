/*
Package metrics adds prometheus metrics and registers a metrics handler on the
/metrics endpoint.

The package is only imported for the side effect of registering its HTTP
handler and adding the included metrics. To use it this way, link this package
into your program:

	import _ "github.com/skuid/spec/metrics"

When not using the default multiplexer in the http.ListenAndServe function
call, the promhttp.Handler() function will need to be invoked manually.
*/
package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/skuid/spec/version"
)

var (
	versionGuauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "version_info",
			Help: "The current git commit and golang version.",
		},
		[]string{"commit", "golang_version"},
	)
)

func init() {
	versionGuauge.WithLabelValues(version.Commit, version.GoVersion).Set(1)
	prometheus.MustRegister(versionGuauge)
	http.Handle("/metrics", promhttp.Handler())
}
