package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// InitLogger initializes the global logger
func InitLogger(debug bool, logPath string) error {
	var config zap.Config

	noLogs := len(logPath) == 0

	if debug {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	} else {
		config = zap.NewProductionConfig()
	}

	if noLogs {
		config.Level = zap.NewAtomicLevelAt(zapcore.FatalLevel)
	}

	if noLogs {
		config.OutputPaths = []string{"stdout"}
		config.ErrorOutputPaths = []string{"stderr"}
	} else {
		config.OutputPaths = []string{logPath}
		config.ErrorOutputPaths = []string{logPath}
	}

	logger, err := config.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return err
	}

	zap.ReplaceGlobals(logger)

	zap.S().Infow("Logger initialized",
		"debug", debug,
		"log_path", logPath)

	return nil
}

// Sync flushes any buffered log entries
func Sync() error {
	return zap.L().Sync()
}
