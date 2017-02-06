/*
Package spec provides a Logger for applications to use and write logs with.

The Logger can be used like so:

	import "github.com/skuid/spec"
	import "github.com/uber-go/zap"

	var logger spec.Logger

	func main() {
		logger.Debug("A debug message")
		logger.Info("An info message")
		logger.Info(
			"An info message with values",
			zap.String("key", "value"),
		)
		logger.Error("An error message")

		err := errors.New("some error")
		logger.Error("An error message", zap.Error(err))
	}

*/
package spec

import (
	"github.com/uber-go/zap"
)

// Logger is a zap.Logger that is initialized with the proper field names and
// time format for Skuid
var Logger zap.Logger

func init() {
	Logger = zap.New(
		zap.NewJSONEncoder(
			zap.RFC3339Formatter("timestamp"),
			zap.MessageKey("message"),
			zap.LevelString("level"),
			// TODO: Write a stacktrace key formatter for zap
		),
	)
}
