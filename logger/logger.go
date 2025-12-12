// This package provides logger that prints logs on standard output
package logger

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/orandin/lumberjackrus"
	"github.com/sirupsen/logrus"
)

var (
	logger      *logrus.Logger
	loggerOnce  sync.Once
	logDir      = getEnv("LOG_DIR", "/usr/src/data")
	logFilePath string
)

func Info(v ...interface{}) {
	getLogger().Info(v...)
}

func Infof(format string, v ...interface{}) {
	getLogger().Infof(format, v...)
}

func Warn(v ...interface{}) {
	getLogger().Warn(v...)
}

func Warnf(format string, v ...interface{}) {
	getLogger().Warnf(format, v...)
}

func Debug(v ...interface{}) {
	getLogger().Debug(v...)
}

func Debugf(format string, v ...interface{}) {
	getLogger().Debugf(format, v...)
}

func Error(v ...interface{}) {
	getLogger().Error(v...)
}

func Errorf(format string, v ...interface{}) {
	getLogger().Errorf(format, v...)
}

func Panic(v ...interface{}) {
	getLogger().Panic(v...)
}

func Panicf(format string, v ...interface{}) {
	getLogger().Panicf(format, v...)
}

func Fatal(v ...interface{}) {
	getLogger().Fatal(v...)
}

func Fatalf(format string, v ...interface{}) {
	getLogger().Fatalf(format, v...)
}

func WithContex(ctx context.Context) *logrus.Entry {
	rid := ""
	if ctx != nil {
		if v := ctx.Value("Request-Id"); v != nil {
			rid, _ = v.(string)
		}
	}
	return getLogger().WithField("request-id", rid)
}

func getLogger() *logrus.Logger {
	loggerOnce.Do(func() {
		if _, err := os.Stat(logDir); os.IsNotExist(err) {
			os.MkdirAll(logDir, os.ModePerm)
		} else if err != nil {
			panic(err)
		}

		logFilePath = filepath.Join(logDir, "compost-bin.log")

		logger = logrus.New()
		logger.SetOutput(os.Stdout)
		logger.SetLevel(logrus.InfoLevel)
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006/01/02 15:04:05.000",
		})
		logger.AddHook(ratateHook())
	})
	return logger
}

func ratateHook() logrus.Hook {
	hook, err := lumberjackrus.NewHook(&lumberjackrus.LogFile{
		Filename:   logFilePath,
		MaxAge:     1,
		MaxSize:    1 << 6,
		MaxBackups: 50,
		Compress:   true,
		LocalTime:  true,
	},
		logrus.InfoLevel,
		&logrus.JSONFormatter{TimestampFormat: "2006/01/02 15:04:05.000"},
		&lumberjackrus.LogFileOpts{})
	if err != nil {
		panic(err)
	}

	return hook
}

func getEnv(variableName, defaultValue string) string {
	if value := os.Getenv(variableName); value != "" {
		return value
	}
	return defaultValue
}
