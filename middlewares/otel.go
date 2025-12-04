package middlewares

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	logsdk "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// https://opentelemetry.io/docs/languages/go/exporters/
// https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp#example-package
// https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp#example-package
// https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp
// https://github.com/open-telemetry/opentelemetry-go-contrib/blob/main/examples/prometheus/main.go

var severityMap = map[string]log.Severity{
	"debug": log.SeverityDebug,
	"info":  log.SeverityInfo,
	"warn":  log.SeverityWarn,
	"error": log.SeverityError,
	"fatal": log.SeverityFatal,
}

// SetupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func SetupOTelSDK(ctx context.Context, traceEndpoint, metricEndpoint, logEndpoint string) (func(context.Context) error, error) {
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
	// Set up trace provider.
	texp, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(traceEndpoint))
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
		mexp, err := otlpmetrichttp.New(ctx, otlpmetrichttp.WithEndpointURL(metricEndpoint))
		if err != nil {
			handleErr(err)
			return shutdown, err
		}
		meterProvider = metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(mexp)))
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	// Set up logger provider.
	lexp, err := otlploghttp.New(ctx, otlploghttp.WithEndpointURL(logEndpoint))
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	processor := logsdk.NewBatchProcessor(lexp)
	loggerProvider := logsdk.NewLoggerProvider(logsdk.WithProcessor(processor))
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

type OpenTelemetryWriter struct {
	ctx    context.Context
	logger log.Logger
}

func (o OpenTelemetryWriter) Write(p []byte) (n int, err error) {
	var msg logMsg
	if err = json.Unmarshal(p, &msg); err != nil {
		return
	}
	n = len(p)

	var timestamp time.Time
	timestamp, err = time.Parse(iso8601, msg.Timestamp)
	if err != nil {
		return
	}

	tags := msg.Tags
	if tags == nil {
		tags = []string{}
	}
	tagValues := make([]log.Value, len(tags))
	for i, v := range tags {
		tagValues[i] = log.StringValue(v)
	}
	tagValue := log.SliceValue(tagValues...)

	severity := log.SeverityDebug
	severityText := "debug"
	if sm, ok := severityMap[msg.Level]; ok {
		severity = sm
		severityText = msg.Level
	}

	record := log.Record{}
	record.SetTimestamp(timestamp)
	record.SetEventName(msg.Name)
	record.SetSeverity(severity)
	record.SetSeverityText(severityText)
	record.AddAttributes(log.KeyValue{
		Key:   "tags",
		Value: tagValue,
	})
	record.SetBody(log.StringValue(msg.Text()))

	o.logger.Emit(o.ctx, record)
	return
}

// OpenTelemetryEventLogger will take an already-constructed zap.Logger, context, and an
// opentelemetry logger and return a Logger that will also "tee" its output to OpenTelemetry.
func OpenTelemetryEventLogger(l *zap.Logger, ctx context.Context, logger log.Logger, level zapcore.Level) *zap.Logger {
	// https://godoc.org/go.uber.org/zap#hdr-Extending_Zap
	stdZapConfig := NewStandardZapConfig()
	opts := zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		enc := stdZapConfig.EncoderConfig
		ddw := zapcore.AddSync(OpenTelemetryWriter{
			ctx,
			logger,
		})
		datadogCore := zapcore.NewCore(zapcore.NewJSONEncoder(enc), ddw, level)
		return zapcore.NewTee(c, datadogCore)
	})

	return l.WithOptions(opts)
}
