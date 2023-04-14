package middlewares

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logLevels = map[string]zapcore.Level{
	"debug":   zap.DebugLevel,
	"info":    zap.InfoLevel,
	"warning": zap.WarnLevel,
	"error":   zap.ErrorLevel,
	"dpanic":  zap.DPanicLevel,
	"panic":   zap.PanicLevel,
	"fatal":   zap.FatalLevel,
}

// GetLogLevel get a zapcore level from an string
//
// levels can be debug, info, warning, error, dpanic, panic, fatal
func GetLogLevel(level string) zapcore.Level {
	if v, ok := logLevels[level]; ok {
		return v
	}

	return zap.ErrorLevel
}
