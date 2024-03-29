/*
Package spec provides a zap.Logger creation function NewStandardLogger() for applications to use and write logs with.

NewStandardLogger() can be used like so:

	package main

	import "go.uber.org/zap"
	import "github.com/skuid/spec"

	func init() {
	    l, err := spec.NewStandardLogger() // handle error
	    zap.ReplaceGlobals(l)
	}

	func main() {
	    zap.L().Debug("A debug message")
	    zap.L().Info("An info message")
	    zap.L().Info(
	        "An info message with values",
	        zap.String("key", "value"),
	    )
	    zap.L().Error("An error message")

	    err := errors.New("some error")
	    zap.L().Error("An error message", zap.Error(err))
	}
*/
package spec

import (
	"flag"

	"github.com/skuid/spec/middlewares"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LevelPflag returns a *pflag.Flag and a *zapcore.Level for the given
// flag arguments
//
//	fset := pflag.CommandLine
//	lflag, level := spec.LevelPflag("level", zapcore.InfoLevel, "Log level")
//	fset.AddFlag(lflag)
func LevelPflag(name string, defaultLevel zapcore.Level, usage string) (*pflag.Flag, *zapcore.Level) {
	lvl := defaultLevel
	set := flag.NewFlagSet("temp", flag.ExitOnError)
	set.Var(&lvl, name, usage)
	return pflag.PFlagFromGoFlag(set.Lookup(name)), &lvl
}

// LevelPflagCommandLine returns a *zapcore.Level for the given flag arguments. The flag is
// added to the pflag.CommandLine flagset
//
//	level := spec.LevelPflagCommandLine("level", zapcore.InfoLevel, "Log level")
func LevelPflagCommandLine(name string, defaultLevel zapcore.Level, usage string) *zapcore.Level {
	fset := pflag.CommandLine
	lflag, level := LevelPflag(name, defaultLevel, usage)
	fset.AddFlag(lflag)
	return level
}

// LevelPflagP returns a *pflag.Flag and a *zapcore.Level for the given
// flag arguments
//
//	fset := pflag.CommandLine
//	lflag, level := spec.LevelPflagP("level","l", zapcore.InfoLevel, "Log level")
//	fset.AddFlag(lflag)
func LevelPflagP(name, shorthand string, defaultLevel zapcore.Level, usage string) (*pflag.Flag, *zapcore.Level) {
	lvl := defaultLevel
	set := flag.NewFlagSet("temp", flag.ExitOnError)
	set.Var(&lvl, name, usage)
	response := pflag.PFlagFromGoFlag(set.Lookup(name))
	response.Shorthand = shorthand
	return response, &lvl
}

// LevelPflagPCommandLine returns a *zapcore.Level for the given flag arguments. The flag is
// added to the pflag.CommandLine flagset
//
//	level := spec.LevelPflagPCommandLine("level", "l", zapcore.InfoLevel, "Log level")
func LevelPflagPCommandLine(name, shorthand string, defaultLevel zapcore.Level, usage string) *zapcore.Level {
	fset := pflag.CommandLine
	lflag, level := LevelPflagP(name, shorthand, defaultLevel, usage)
	fset.AddFlag(lflag)
	return level
}

// NewStandardLevelLogger creates a new zap.Logger based on common
// configuration. It accepts a zapcore.Level as the level to filter logs on.
//
// This is intended to be used with zap.ReplaceGlobals() in an application's
// main.go.
func NewStandardLevelLogger(level zapcore.Level) (l *zap.Logger, err error) {
	config := middlewares.NewStandardZapLevelConfig(level)
	return config.Build()
}

// NewStandardLogger creates a new zap.Logger based on common configuration
//
// This is intended to be used with zap.ReplaceGlobals() in an application's
// main.go.
func NewStandardLogger() (l *zap.Logger, err error) {
	config := middlewares.NewStandardZapConfig()
	return config.Build()
}

// GetLogLevel get a zapcore level from an string
//
// levels can be debug, info, warning, error, dpanic, panic, fatal
func GetLogLevel(level string) zapcore.Level {
	return middlewares.GetLogLevel(level)
}
