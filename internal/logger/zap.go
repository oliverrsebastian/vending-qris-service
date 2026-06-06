package logger

import (
	"fmt"

	"go.uber.org/zap"
)

var logger *zap.Logger

func Init() (sync func(), err error) {
	logger, err = zap.NewProduction()
	if err != nil {
		return nil, err
	}

	return func() {
		logger.Sync()
	}, nil
}

func Get() *zap.Logger {
	return logger
}

func Info(msg string, args ...interface{}) {
	logger.Info(fmt.Sprintf(msg, args...))
}

func Warn(msg string, args ...interface{}) {
	logger.Warn(fmt.Sprintf(msg, args...))
}

func Debug(msg string, args ...interface{}) {
	logger.Debug(fmt.Sprintf(msg, args...))
}

func Fatal(msg string, args ...interface{}) {
	logger.Fatal(fmt.Sprintf(msg, args...))
}

func Error(msg string, args ...interface{}) {
	logger.Error(fmt.Sprintf(msg, args...))
}
