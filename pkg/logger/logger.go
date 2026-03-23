package logger

import (
	"fmt"
	"strings"

	"gin-rocket/pkg/configx"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(cfg configx.LogConfig, env string) (*zap.Logger, error) {
	level := zap.NewAtomicLevel()
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return nil, fmt.Errorf("parse log level: %w", err)
	}

	encoding := cfg.Format
	if encoding == "" {
		encoding = "json"
	}

	zapCfg := zap.Config{
		Level:       level,
		Development: strings.EqualFold(env, "dev") || strings.EqualFold(env, "local"),
		Encoding:    encoding,
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
		DisableStacktrace: cfg.DisableStacktrace,
	}

	return zapCfg.Build(zap.AddCaller())
}
