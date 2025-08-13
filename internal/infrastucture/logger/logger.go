package log

import (
	"context"
	"os"

	"github.com/google/wire"
	"github.com/th1enq/ViettelSMS_ServerService/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var LoggerSet = wire.NewSet(LoadLogger)

func LoadLogger(cfg *config.Config) (*zap.Logger, error) {
	level, err := zapcore.ParseLevel(cfg.Logger.Level)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll("./logs", 0755); err != nil {
		return nil, err
	}

	lumberjackLogger := &lumberjack.Logger{
		Filename:   cfg.Logger.FilePath,
		MaxSize:    cfg.Logger.MaxSize,
		MaxAge:     cfg.Logger.MaxAge,
		MaxBackups: cfg.Logger.MaxBackups,
		Compress:   true,
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(lumberjackLogger),
		level,
	)

	consoleCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		level,
	)

	core := zapcore.NewTee(fileCore, consoleCore)

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger, nil
}

func getZapLoggerLevel(level string) zap.AtomicLevel {
	switch level {
	case "debug":
		return zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		return zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		return zap.NewAtomicLevelAt(zap.ErrorLevel)
	case "panic":
		return zap.NewAtomicLevelAt(zap.PanicLevel)
	default:
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	}
}

func LoggerWithContext(_ context.Context, logger *zap.Logger) *zap.Logger {
	//TODO: Add request ID to context
	return logger
}
