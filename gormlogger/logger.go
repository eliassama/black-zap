package gormlogger

import (
	"context"
	"errors"
	"fmt"
	"github.com/eliassama/black-zap/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Log gorm 日志结构
type Log struct {
	ZapLogger                 *zap.Logger
	LogLevel                  gormLogger.LogLevel
	SlowThreshold             time.Duration
	SkipCallerLookup          bool
	IgnoreRecordNotFoundError bool
}

// New 创建 gorm 日志实例
func New(level gormLogger.LogLevel, configs ...*logger.Conf) Log {

	log := Log{
		ZapLogger:                 logger.New("dataBase", configs...),
		LogLevel:                  level,
		SlowThreshold:             3 * time.Second,
		SkipCallerLookup:          false,
		IgnoreRecordNotFoundError: true,
	}

	log.SetAsDefault()
	log.LogMode(level)

	return log
}

// SetAsDefault 设置默认日志
func (l Log) SetAsDefault() {
	gormLogger.Default = l
}

// LogMode 设置日志等级
func (l Log) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	return Log{
		ZapLogger:                 l.ZapLogger,
		SlowThreshold:             l.SlowThreshold,
		LogLevel:                  level,
		SkipCallerLookup:          l.SkipCallerLookup,
		IgnoreRecordNotFoundError: l.IgnoreRecordNotFoundError,
	}
}

// Info 正常日志打印
func (l Log) Info(_ context.Context, str string, args ...interface{}) {
	if l.LogLevel >= gormLogger.Info {
		l.logger().Sugar().Infof(fmt.Sprintf("[%s] %s", utils.FileWithLineNum(), str), args...)
	}
}

// Warn 告警日志打印
func (l Log) Warn(_ context.Context, str string, args ...interface{}) {
	if l.LogLevel >= gormLogger.Warn {
		l.logger().Sugar().Warnf(fmt.Sprintf("[%s] %s", utils.FileWithLineNum(), str), args...)
	}
}

// Error 异常日志打印
func (l Log) Error(_ context.Context, str string, args ...interface{}) {
	if l.LogLevel >= gormLogger.Error {
		l.logger().Sugar().Errorf(fmt.Sprintf("[%s] %s", utils.FileWithLineNum(), str), args...)
	}
}

// Trace SQL 日志打印
func (l Log) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= gormLogger.Silent {
		return
	}

	elapsed := time.Since(begin)

	sql, rows := fc()
	if rows == -1 {
		rows = 0
	}

	switch {
	case err != nil && l.LogLevel >= gormLogger.Error && (!errors.Is(err, gorm.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		l.logger().Error("trace",
			zap.Error(err),
			zap.String("fileWithLineNum", utils.FileWithLineNum()),
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)

	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= gormLogger.Warn:
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)

		l.logger().Warn("trace",
			zap.Error(err),
			zap.String("fileWithLineNum", utils.FileWithLineNum()),
			zap.String("slowLog", slowLog),
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)

	case l.LogLevel == gormLogger.Info:
		l.logger().Info("trace",
			zap.String("fileWithLineNum", utils.FileWithLineNum()),
			zap.Duration("elapsed", elapsed),
			zap.Int64("rows", rows),
			zap.String("sql", sql),
		)
	}
}

var (
	gormPackage    = filepath.Join("gorm.io", "gorm")
	zapGormPackage = filepath.Join("moul.io", "zapgorm2")
)

func (l Log) logger() *zap.Logger {
	for i := 2; i < 15; i++ {
		_, file, _, ok := runtime.Caller(i)
		switch {
		case !ok:
		case strings.HasSuffix(file, "_test.go"):
		case strings.Contains(file, gormPackage):
		case strings.Contains(file, zapGormPackage):
		default:
			return l.ZapLogger.WithOptions(zap.AddCallerSkip(i))
		}
	}
	return l.ZapLogger
}
