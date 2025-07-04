package logger

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a new zap.Logger instance.
// The 'env' parameter can be "development" or "production".
func New(env string) *zap.Logger {
	var logger *zap.Logger
	var err error

	if env == "production" {
		// Production logger: JSON format, info level and above.
		config := zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		logger, err = config.Build()
	} else {
		// Development logger: Human-readable console format.
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		logger, err = config.Build()
	}

	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}

	// Replace the global logger with our new zap logger.
	// This is useful for third-party libraries that use the standard log package.
	zap.ReplaceGlobals(logger)

	return logger
}
