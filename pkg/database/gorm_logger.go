package database

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	gormlogger "gorm.io/gorm/logger"
)

type zapGormLogger struct {
	logger        *zap.Logger
	level         gormlogger.LogLevel
	slowThreshold time.Duration
}

func newGormLogger(log *zap.Logger, level string, slowThreshold time.Duration) gormlogger.Interface {
	return &zapGormLogger{
		logger:        log.Named("gorm"),
		level:         parseGormLogLevel(level),
		slowThreshold: slowThreshold,
	}
}

func parseGormLogLevel(level string) gormlogger.LogLevel {
	switch strings.ToLower(level) {
	case "silent":
		return gormlogger.Silent
	case "error":
		return gormlogger.Error
	case "info":
		return gormlogger.Info
	default:
		return gormlogger.Warn
	}
}

func (l *zapGormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	cloned := *l
	cloned.level = level
	return &cloned
}

func (l *zapGormLogger) Info(_ context.Context, msg string, data ...any) {
	if l.level < gormlogger.Info {
		return
	}

	l.logger.Sugar().Infof(msg, data...)
}

func (l *zapGormLogger) Warn(_ context.Context, msg string, data ...any) {
	if l.level < gormlogger.Warn {
		return
	}

	l.logger.Sugar().Warnf(msg, data...)
}

func (l *zapGormLogger) Error(_ context.Context, msg string, data ...any) {
	if l.level < gormlogger.Error {
		return
	}

	l.logger.Sugar().Errorf(msg, data...)
}

func (l *zapGormLogger) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)

	switch {
	case err != nil && l.level >= gormlogger.Error && !errors.Is(err, gormlogger.ErrRecordNotFound):
		sql, rows := fc()
		l.logger.Error(
			"gorm trace",
			zap.Error(err),
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
	case l.slowThreshold != 0 && elapsed > l.slowThreshold && l.level >= gormlogger.Warn:
		sql, rows := fc()
		l.logger.Warn(
			"gorm slow query",
			zap.String("message", fmt.Sprintf("slow query >= %s", l.slowThreshold)),
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
	case l.level >= gormlogger.Info:
		sql, rows := fc()
		l.logger.Info(
			"gorm trace",
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
	}
}
