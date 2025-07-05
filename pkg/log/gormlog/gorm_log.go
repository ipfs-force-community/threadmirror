package gormlog

import (
	"log/slog"

	slogGorm "github.com/orandin/slog-gorm"
	gormLogger "gorm.io/gorm/logger"
)

func New(logger *slog.Logger) gormLogger.Interface {
	return slogGorm.New(slogGorm.WithHandler(logger.Handler()))
}
