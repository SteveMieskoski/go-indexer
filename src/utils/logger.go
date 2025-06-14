package utils

import (
	"go.uber.org/zap"
)

var Logger *zap.SugaredLogger

func InitializeLogger() {
	logger1, _ := zap.NewProduction()
	defer logger1.Sync() // flushes buffer, if any
	Logger = logger1.Sugar()
}
