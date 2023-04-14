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

func GetLogLevel(level string) zapcore.Level {
	if v, ok := logLevels[level]; ok {
		return v
	}

	return zap.ErrorLevel
}
