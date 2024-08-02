package gormlogger

import (
	"context"
	"testing"

	gormLogger "gorm.io/gorm/logger"
)

func TestGormLoggerNew(t *testing.T) {
	t.Log("GormLoggerNew Test Start")

	log := New(gormLogger.Info)
	log.Info(context.Background(), "GormLoggerNew Test Log")
	log.Warn(context.Background(), "GormLoggerNew Test Log")
	log.Error(context.Background(), "GormLoggerNew Test Log")
	t.Log("GormLoggerNew Test Done")
}
