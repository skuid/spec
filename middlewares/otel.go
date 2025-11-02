package middlewares

import (
	"context"
	"errors"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

// SetupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func SetupOTelSDK(ctx context.Context) (func(context.Context) error, error) {
	var shutdownFuncs []func(context.Context) error
	var err error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// https://opentelemetry.io/docs/languages/go/exporters/
	// https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp#example-package
	// https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc#example-package
	// https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp
	// https://github.com/open-telemetry/opentelemetry-go-contrib/blob/main/examples/prometheus/main.go
	// TODO: allow endpoint config via OTEL_EXPORTER_OTLP_ENDPOINT, or from ctx, or both?
	// TODO: modify middlewares.Logging to use the global.SetLoggerProvider from below via global.GetLoggerProvider

	// Set up trace provider.
	texp, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint("localhost:4317"))
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	tracerProvider := trace.NewTracerProvider(trace.WithBatcher(texp))
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// Set up meter provider.
	var meterProvider *metric.MeterProvider
	if os.Getenv("PROMETHEUS_METRICS") == "true" {
		exporter, err := prometheus.New()
		if err != nil {
			handleErr(err)
			return shutdown, err
		}
		meterProvider = metric.NewMeterProvider(metric.WithReader(exporter))
	} else {
		mexp, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithEndpoint("localhost:4317"))
		if err != nil {
			handleErr(err)
			return shutdown, err
		}
		meterProvider = metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(mexp)))
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	// Set up logger provider.
	lexp, err := otlploghttp.New(ctx, otlploghttp.WithEndpoint("localhost:4317"))
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	processor := log.NewBatchProcessor(lexp)
	loggerProvider := log.NewLoggerProvider(log.WithProcessor(processor))
	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)

	return shutdown, err
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}
