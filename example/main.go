/*
Package example contains an example web application that uses packages
contained in this project.
*/
package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/skuid/spec"
	"github.com/skuid/spec/lifecycle"
	_ "github.com/skuid/spec/metrics"
	"github.com/skuid/spec/middlewares"
	"github.com/uber-go/zap"
)

func init() {
	rand.Seed(int64(time.Now().Second()))
}

// A default handler
func hello(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "hello!"}`))
}

// A function that returns an error
func barf(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(`{"message": "stop it"}`))
}

// A function that takes a variable amount of time
func random(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	x := rand.Int() % 6000
	time.Sleep(time.Duration(int64(x)) * time.Millisecond)

	w.Write([]byte(fmt.Sprintf(`{"slept": %d}`, x)))
}

// flip sets lifecycle.Ready to the inverse of it's current state
func flip(w http.ResponseWriter, r *http.Request) {
	if lifecycle.Ready {
		lifecycle.Ready = false
	} else {
		lifecycle.Ready = true
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"Ready": "%t"}`, lifecycle.Ready)))
}

func main() {

	mux := http.NewServeMux()

	mux.HandleFunc("/hello", hello)
	mux.HandleFunc("/barf", barf)
	mux.HandleFunc("/flip", flip)
	mux.HandleFunc("/random", random)

	handler := middlewares.Apply(
		mux,
		middlewares.InstrumentRoute(),
		middlewares.Logging(),
		middlewares.AccessControlAllowOrigin("*"),
		middlewares.AddHeaders(map[string]string{"X-Frame-Options": "DENY"}),
	)

	internalMux := http.NewServeMux()
	internalMux.Handle("/", handler)
	internalMux.Handle("/metrics", promhttp.Handler())
	internalMux.HandleFunc("/live", lifecycle.LivenessHandler)
	internalMux.HandleFunc("/ready", lifecycle.ReadinessHandler)

	hostPort := ":3000"

	spec.Logger.Info("Server is starting", zap.String("listen", hostPort))

	server := &http.Server{Addr: hostPort, Handler: internalMux}
	lifecycle.ShutdownOnTerm(server)

	err := server.ListenAndServe()
	if err != http.ErrServerClosed {
		spec.Logger.Fatal("", zap.Error(err))
	}
	spec.Logger.Info("Server gracefully stopped")
}
