package logger

import (
	"testing"
)

func TestLoggerNew(t *testing.T) {
	t.Log("LoggerNew Test Start")

	log := NewDebug("test")
	log.Debug("LoggerNew Test Debug Logger")
	log.Info("LoggerNew Test Debug Logger")
	log.Warn("LoggerNew Test Debug Logger")
	log.Error("LoggerNew Test Debug Logger")

	warnLog := NewWarn("test")
	warnLog.Debug("LoggerNew Test Warn Logger")
	warnLog.Info("LoggerNew Test Warn Logger")
	warnLog.Warn("LoggerNew Test Warn Logger")
	warnLog.Error("LoggerNew Test Warn Logger")

	errorLog := NewError("test")
	errorLog.Debug("LoggerNew Test Error Logger")
	errorLog.Info("LoggerNew Test Error Logger")
	errorLog.Warn("LoggerNew Test Error Logger")
	errorLog.Error("LoggerNew Test Error Logger")

	t.Log("LoggerNew Test Done")
}
