package logger

import (
	"context"
	"regexp"
	"time"

	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

type DatabaseLogger struct {
	LogLevel gormlogger.LogLevel
}

func NewDatabaseLogger() *DatabaseLogger {
	return &DatabaseLogger{LogLevel: gormlogger.Info}
}

func (l *DatabaseLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	logger := *l
	logger.LogLevel = level
	return &logger
}

func (l *DatabaseLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Info {
		Infof(msg, data...)
	}
}

func (l *DatabaseLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Warn {
		Warnf(msg, data...)
	}
}

func (l *DatabaseLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Error {
		Errorf(msg, data...)
	}
}

func (l *DatabaseLogger) Trace(
	ctx context.Context,
	begin time.Time,
	fc func() (sql string, rowsAffected int64),
	err error,
) {
	if l.LogLevel <= gormlogger.Silent {
		return
	}

	rid := getRequestId(ctx)
	elapsed := time.Since(begin)
	sql, rows := fc()
	if containsSensitiveWord(sql) {
		pattern := regexp.MustCompile(`'[^']*`)
		sql = pattern.ReplaceAllString(sql, "'***Sensitive***'")
	}

	logFields := map[string]interface{}{
		"request-id": rid,
		"duration":   elapsed.String(),
		"rows":       rows,
		"files":      utils.FileWithLineNum(),
		"sql":        sql,
	}

	switch {
	case err != nil && l.LogLevel >= gormlogger.Error:
		logFields["error"] = err
		getLogger().WithFields(logFields).Error("SQL ERROR")
	case elapsed >= 200*time.Millisecond && l.LogLevel >= gormlogger.Warn:
		getLogger().WithFields(logFields).Warn("SLOW SQL >= 200ms")
	default:
		if l.LogLevel >= gormlogger.Info {
			getLogger().WithFields(logFields).Info("SQL EXECUTED")
		}
	}
}

func getRequestId(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if v := ctx.Value("Request-Id"); v != nil {
		if rid, ok := v.(string); ok {
			return rid
		}
	}
	return ""
}
