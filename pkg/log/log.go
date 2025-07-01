package log

import (
	"fmt"
	"log/slog"

	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
	"go.uber.org/zap/zapcore"
)

// Logger provides structured logging interface
type Logger struct {
	*slog.Logger
	zapLogger *zap.Logger
}

// New creates a new logger instance
func New(level string, devMode bool) (*Logger, error) {
	var (
		zapLogger *zap.Logger
		err       error
		zapConfig zap.Config
		zapLevel  zapcore.Level
	)

	if devMode {
		zapConfig = zap.NewDevelopmentConfig()
		// 设置为 console 编码格式，不使用 JSON
		zapConfig.Encoding = "console"
		// 配置优美的输出格式
		zapConfig.EncoderConfig = zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalColorLevelEncoder, // 带颜色的级别
			EncodeTime:     zapcore.ISO8601TimeEncoder,       // 可读的时间格式
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder, // 简短的调用者信息
		}
	} else {
		// Use production defaults but render as human-readable console instead of JSON.
		zapConfig = zap.NewProductionConfig()
		zapConfig.Encoding = "console"
		zapConfig.EncoderConfig = zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder, // No color in production
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}
	}

	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		return nil, fmt.Errorf("invalid log level %s: %w", level, err)
	}

	zapConfig.Level.SetLevel(zapLevel)
	zapLogger, err = zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("build zap logger: %w", err)
	}

	// Create slog handler from zap logger
	handler := zapslog.NewHandler(zapLogger.Core())

	slogLogger := slog.New(handler)

	return &Logger{
		Logger:    slogLogger,
		zapLogger: zapLogger,
	}, nil
}

// Close flushes any buffered log entries
func (l *Logger) Close() error {
	if l.zapLogger != nil {
		return l.zapLogger.Sync()
	}
	return nil
}
