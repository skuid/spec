package middlewares

import (
	"github.com/DataDog/datadog-go/statsd"
)

var statsdClient *statsd.Client

// InitClient accepts config parameters and builds the dogstatsd Client, accessible by the getter Client()
// An error is returned only if a problem is encountered setting up the Client.
func InitClient(host string, port string, prefix string) error {
	var err error
	statsdClient, err = statsd.New(host + ":" + port)
	if err != nil {
		return err
	}
	statsdClient.Namespace = prefix
	return nil
}

// Client returns a pointer to the dogstatsd Client, useful for logging metrics
func Client() *statsd.Client {
	return statsdClient
}
