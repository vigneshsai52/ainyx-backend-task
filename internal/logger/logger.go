// Package logger wraps Uber Zap and exposes a package-level sugar logger.
package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

// Init initialises the global zap logger.
// Call once from main before starting the server.
func Init() {
	env := os.Getenv("APP_ENV")

	var cfg zap.Config
	if env == "production" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	var err error
	log, err = cfg.Build()
	if err != nil {
		panic("failed to initialise logger: " + err.Error())
	}
}

// Get returns the global sugared logger.
// Panics if Init has not been called first.
func Get() *zap.SugaredLogger {
	if log == nil {
		panic("logger: Init() must be called before Get()")
	}
	return log.Sugar()
}

// Sync flushes any buffered log entries. Call on shutdown.
func Sync() {
	if log != nil {
		_ = log.Sync()
	}
}
