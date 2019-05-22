package middlewares

import (
	"encoding/json"
	"os"
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

// Hardcoded maps and values used in DataDogWriter.Write() for constructing a statsd Event from a log message buffer.
var eventMap = map[string]statsd.EventAlertType{
	"debug":   statsd.Success,
	"info":    statsd.Info,
	"warning": statsd.Warning,
	"error":   statsd.Error,
	"fatal":   statsd.Error,
}
var iso8601 = "2006-01-02T15:04:05.000Z0700"

type logMsg struct {
	// --- explicit log message fields
	Caller     string   `json:"caller"`
	Level      string   `json:"level"`
	Message    string   `json:"message"`
	Name       string   `json:"name"`
	Stacktrace string   `json:"stacktrace"`
	Timestamp  string   `json:"timestamp"`
	Tags       []string `json:"tags"`
	// --- HTTP request fields added by middlewares.Logging()
	Path       string `json:"string"`
	Method     string `json:"method"`
	Status     int    `json:"status"`
	Query      string `json:"query"`
	RemoteAddr string `json:"remote_addr"`
	UserAgent  string `json:"user_agent"`
	BodyBytes  int    `json:"body_bytes"`
	UserID     string `json:"userId"`
	SiteID     string `json:"siteId"`
}

func (msg logMsg) Text() string {
	text, err := json.Marshal(msg)
	if err != nil {
		return "Error: log message could not be stringified"
	}
	return string(text)
}

// DataDogWriter implements io.Writer. It should be made into a [WriteSyncer](https://godoc.org/go.uber.org/zap/zapcore#WriteSyncer)
// for sending Zap logs to DataDog as [Events](https://godoc.org/github.com/DataDog/datadog-go/statsd#Event), using
// datadog-go/statsd [Client.Event()](https://godoc.org/github.com/DataDog/datadog-go/statsd#Client.Event).
type DataDogWriter struct {
	client *statsd.Client
}
// Write will take a JSON-formatted log message from zap.L().[Level]() as a byte slice, format it, and send it as
// an Event to d.client.Event().
// @param p []byte - the byte array of the buffered log message.
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
	if msg.Name == "" {
		// Event.Title is required, so force msg.Name to be non-empty. os.Args[0] is the executable name.
		msg.Name = os.Args[0] + " event"
	}
	evt := statsd.Event{
		AlertType: eventMap[msg.Level],
		Tags: tags,
		Title: msg.Name,
		Text: msg.Text(),
		Timestamp: ts,
	}

	if msg.Level == "debug" {
		evt.Priority = statsd.Low
	} else {
		evt.Priority = statsd.Normal
	}

	if err := d.client.Event(&evt); err != nil {
		return 0, err
	}
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