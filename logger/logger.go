package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// InitLogger initializes the global logger
func InitLogger(debug bool, noLogs bool, logPath string) error {
	var config zap.Config

	// ベースとなる設定を選択
	if debug {
		// 開発環境用の設定
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		// STDIOのリクエストとレスポンスもログ出力するためにDebugレベルを設定
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	} else {
		// 本番環境用の設定
		config = zap.NewProductionConfig()
	}

	// noLogsが指定されている場合はログレベルをFatalのみに設定
	if noLogs {
		config.Level = zap.NewAtomicLevelAt(zapcore.FatalLevel)
	}

	// 出力先の設定
	if noLogs {
		if logPath != "" {
			// noLogs && len(logPath) > 0: logPathにのみ出力
			config.OutputPaths = []string{logPath}
			config.ErrorOutputPaths = []string{logPath}
		} else {
			// noLogs && len(logPath) == 0: どこにも出力しない
			config.OutputPaths = []string{}
			config.ErrorOutputPaths = []string{}
		}
	} else {
		if logPath != "" {
			// !noLogs && len(logPath) > 0: logPathとstderr両方に出力
			config.OutputPaths = []string{"stdout", logPath}
			config.ErrorOutputPaths = []string{"stderr", logPath}
		} else {
			// !noLogs && len(logPath) == 0: stderrのみに出力
			config.OutputPaths = []string{"stdout"}
			config.ErrorOutputPaths = []string{"stderr"}
		}
	}

	// ロガーを構築
	logger, err := config.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return err
	}

	// グローバルロガーを置き換え
	zap.ReplaceGlobals(logger)

	// SugaredLoggerも初期化
	zap.S().Infow("Logger initialized",
		"debug", debug,
		"no_logs", noLogs,
		"log_path", logPath)

	return nil
}

// Sync flushes any buffered log entries
func Sync() error {
	return zap.L().Sync()
}
