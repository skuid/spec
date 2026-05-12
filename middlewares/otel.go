package middlewares

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	logsdk "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// https://opentelemetry.io/docs/languages/go/resources/
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

// SetupOTelSDK bootstraps the OpenTelemetry pipeline. Equivalent to statsd.go InitClient(...).
// If it does not return an error, make sure to call shutdown for proper cleanup.
func SetupOTelSDK(ctx context.Context, traceEndpoint, metricEndpoint, logEndpoint, serviceName string, env string, globalTags []string) (func(context.Context) error, error) {
	var shutdownFuncs []func(context.Context) error
	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined and each registered cleanup will be invoked once.
	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	var err error
	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	var otelResource *resource.Resource
	otelResource, err = resource.New(
		ctx,
		resource.WithProcess(),
		resource.WithHost(),
		resource.WithContainer(),
		resource.WithAttributes(
			semconv.DeploymentEnvironmentName(env),
			semconv.ServiceName(serviceName),
			attribute.StringSlice("tags", globalTags),
		),
	)
	if err != nil {
		handleErr(err)
		return shutdown, err
	}

	if logEndpoint != "" {
		// Set up logger provider.
		lexp, err := otlploghttp.New(ctx, otlploghttp.WithEndpointURL(logEndpoint))
		if err != nil {
			handleErr(err)
			return shutdown, err
		}
		processor := logsdk.NewBatchProcessor(lexp)
		loggerProvider := logsdk.NewLoggerProvider(
			logsdk.WithProcessor(processor),
			logsdk.WithResource(otelResource),
		)
		shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
		global.SetLoggerProvider(loggerProvider)
	}

	if traceEndpoint != "" {
		// Set up trace provider.
		prop := propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		)
		otel.SetTextMapPropagator(prop)
		texp, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(traceEndpoint))
		if err != nil {
			handleErr(err)
			return shutdown, err
		}
		tracerProvider := trace.NewTracerProvider(
			trace.WithBatcher(texp),
			trace.WithResource(otelResource),
		)
		shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
		otel.SetTracerProvider(tracerProvider)
	}

	if metricEndpoint != "" {
		// Set up meter provider.
		var meterProvider *metric.MeterProvider
		var exporter metric.Reader
		if os.Getenv("PROMETHEUS_METRICS") == "true" {
			exporter, err = prometheus.New()
			if err != nil {
				handleErr(err)
				return shutdown, err
			}
		} else {
			var mexp metric.Exporter
			mexp, err = otlpmetrichttp.New(ctx, otlpmetrichttp.WithEndpointURL(metricEndpoint))
			if err != nil {
				handleErr(err)
				return shutdown, err
			}
			exporter = metric.NewPeriodicReader(mexp)
		}
		meterProvider = metric.NewMeterProvider(
			metric.WithReader(exporter),
			metric.WithResource(otelResource),
		)
		shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
		otel.SetMeterProvider(meterProvider)

		startMeter, err := meterProvider.Meter("server").Int64Counter("server_start")
		if err != nil {
			handleErr(err)
			return shutdown, err
		}
		startMeter.Add(ctx, 1)
	}

	return shutdown, err
}

// OpenTelemetryWriter implements io.Writer to send Zap logs to OpenTelemetry as log records.
type OpenTelemetryWriter struct {
	ctx    context.Context
	logger log.Logger
}

func (o OpenTelemetryWriter) Write(p []byte) (n int, err error) {
	var msg logMsg
	if err = json.Unmarshal(p, &msg); err != nil {
		return
	}

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
	n = len(p)
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
		otelCore := zapcore.NewCore(zapcore.NewJSONEncoder(enc), ddw, level)
		return zapcore.NewTee(c, otelCore)
	})

	return l.WithOptions(opts)
}
