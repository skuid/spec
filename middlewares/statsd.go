package middlewares

import (
	"github.com/DataDog/datadog-go/statsd"
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
