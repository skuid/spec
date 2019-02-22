package middlewares

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var c *statsd.Client

// InitClient accepts config parameters and builds the dogstatsd Client, accessible by the getter Client()
// An error is returned only if a problem is encountered setting up the Client.
func InitClient(host string, port string, prefix string, globalTags []string) error {
	var err error
	c, err = statsd.New(host + ":" + port)
	if err != nil {
		return err
	}
	c.Tags = append(c.Tags, globalTags...)
	c.Namespace = prefix + "."
	c.Incr("server_start", []string(nil), 1)
	return nil
}

// Client returns a pointer to the dogstatsd Client, useful for logging metrics
func Client() *statsd.Client {
	return c
}


var eventMap = map[string]statsd.EventAlertType{
	"debug": statsd.Success,
	"info": statsd.Info,
	"warning": statsd.Warning,
	"error": statsd.Error,
	"fatal": statsd.Error,
}
var iso8601 = "2006-01-02T15:04:05.000Z0700"
type logMsg struct {
	Caller string `json:"caller"`
	Level string `json:"level"`
	Message string `json:"message"`
	Name string `json:"name"`
	Stacktrace string `json:"stacktrace"`
	Timestamp string `json:"timestamp"`
	Tags []string `json:"tags"`
}
type DataDogWriter struct {
	client *statsd.Client
}
func (d DataDogWriter) Write(p []byte) (n int, err error) {
	var msg logMsg
	if err := json.Unmarshal(p, &msg); err != nil {
		return 0, err
	}
	ts, err := time.Parse(iso8601, msg.Timestamp)
	if err != nil {
		return 0, err
	}

	tags := msg.Tags
	if tags == nil {
		tags = []string{}
	}
	if d.client != nil && d.client.Tags != nil {
		tags = append(tags, d.client.Tags...)
	}
	evt := statsd.Event{
		AlertType: eventMap[msg.Level],
		Tags: tags,
		Title: msg.Name,
		Text: fmt.Sprintf("{\"message\": \"%s\", \"caller\": \"%s\", \"stack\": \"%s\"", msg.Message, msg.Caller, msg.Stacktrace),
		Timestamp: ts,
	}

	if msg.Level == "debug" {
		evt.Priority = statsd.Low
	} else {
		evt.Priority = statsd.Normal
	}

	d.client.Event(&evt)
	// n is supposed to be the number of bytes written, must return error if n < len(p)
	return len(p), nil
}

// DataDogEventLogger will take an already-constructed zap.Logger and a datadog statsd.Client
// and return a Logger that will also "tee" its output to DataDog Events.
func DataDogEventLogger(l *zap.Logger, sc *statsd.Client, level zapcore.Level) (*zap.Logger) {
	// https://godoc.org/go.uber.org/zap#hdr-Extending_Zap
	stdZapConfig := NewStandardZapConfig()
	opts := zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		enc := stdZapConfig.EncoderConfig
		ddw := zapcore.AddSync(DataDogWriter{
			sc,
		})
		datadogCore := zapcore.NewCore(zapcore.NewJSONEncoder(enc), ddw, level)
		return zapcore.NewTee(c, datadogCore)
	})

	return l.WithOptions(opts)
}