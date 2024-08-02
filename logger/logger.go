package logger

import (
	"github.com/eliassama/black-zap/report"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// DateTimeEncoder is ...
const DateTimeEncoder = "2006-01-02 15:04:05"

// Level 日志等级
var Level = struct {
	Debug string
	Info  string
	Warn  string
	Error string
}{
	Debug: "debug",
	Info:  "info",
	Warn:  "warn",
	Error: "error",
}

// Type 日志类型
var Type = struct {
	STD    string
	FILE   string
	REPORT string
}{
	STD:    "std",
	FILE:   "file",
	REPORT: "report",
}

// Conf 日志配置
type Conf struct {
	Level    string
	Type     string
	Path     string
	CallBack func(level string, message string)
}

// New 创建默认日志实例
func New(serverName string, configs ...*Conf) *zap.Logger {
	return NewInfo(serverName, configs...)
}

// NewDebug 创建标准输出为 Debug 等级的日志实例
func NewDebug(serverName string, configs ...*Conf) *zap.Logger {
	return create(serverName, append([]*Conf{{
		Level: Level.Debug,
		Type:  Type.STD,
	}}, configs...)...)
}

// NewInfo 创建标准输出为 Info 等级的日志实例
func NewInfo(serverName string, configs ...*Conf) *zap.Logger {
	return create(serverName, append([]*Conf{{
		Level: Level.Info,
		Type:  Type.STD,
	}}, configs...)...)
}

// NewWarn 创建标准输出为 Warn 等级的日志实例
func NewWarn(serverName string, configs ...*Conf) *zap.Logger {
	return create(serverName, append([]*Conf{{
		Level: Level.Warn,
		Type:  Type.STD,
	}}, configs...)...)
}

// NewError 创建标准输出为 Error 等级的日志实例
func NewError(serverName string, configs ...*Conf) *zap.Logger {
	return create(serverName, append([]*Conf{{
		Level: Level.Error,
		Type:  Type.STD,
	}}, configs...)...)
}

// create 创建日志
func create(serverName string, configs ...*Conf) *zap.Logger {
	if serverName == "" {
		serverName = "server"
	}

	// stdConf 日志标准输出配置
	var stdConf *Conf

	// fileConf 日志文件输出配置
	var fileConf *Conf

	// reportConf 日志上报配置
	var reportConf *Conf

	if configs != nil && len(configs) > 0 {
		for _, config := range configs {
			if config == nil {
				continue
			}

			switch config.Type {
			case Type.STD:
				stdConf = config

			case Type.FILE:
				fileConf = config

			case Type.REPORT:
				reportConf = config

			default:
				panic("Invalid Logger Type")
			}

		}
	}

	core := getLogCore(serverName, stdConf, fileConf, reportConf)

	fields := zap.Fields(
		zap.String("serverName", serverName),
	)

	return zap.New(core, zap.AddCaller(), fields)
}

func getEncoderConfig() zapcore.EncoderConfig {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.CallerKey = "codeSite"
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(DateTimeEncoder)
	encoderConfig.EncodeName = zapcore.FullNameEncoder

	return encoderConfig
}

func getFileEncoder() zapcore.Encoder {
	return zapcore.NewJSONEncoder(getEncoderConfig())
}

func getReportEncoder() zapcore.Encoder {
	encoderConfig := getEncoderConfig()

	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder // level大写染色编码器

	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		var encoder strings.Builder
		encoder.WriteString("[")
		encoder.WriteString(t.Format(DateTimeEncoder))
		encoder.WriteString("]")
		enc.AppendString(encoder.String())
	}

	encoderConfig.EncodeCaller = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		var callerStr strings.Builder
		callerStr.WriteString("[")
		callerStr.WriteString(caller.TrimmedPath())
		callerStr.WriteString("]\n")
		enc.AppendString(callerStr.String())
	}

	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getStdEncoder() zapcore.Encoder {
	encoderConfig := getEncoderConfig()

	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // level大写染色编码器

	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		var encoder strings.Builder
		encoder.WriteString("[")
		encoder.WriteString(t.Format(DateTimeEncoder))
		encoder.WriteString("]")
		enc.AppendString(encoder.String())
	}

	encoderConfig.EncodeCaller = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		var callerStr strings.Builder
		callerStr.WriteString("[")
		callerStr.WriteString(caller.TrimmedPath())
		callerStr.WriteString("]")
		enc.AppendString(callerStr.String())
	}

	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogFileName(serverName string, logType string, path string) string {
	var logFileName strings.Builder

	logFileName.WriteString(path)
	logFileName.WriteString(serverName)
	logFileName.WriteString(".")
	logFileName.WriteString(logType)
	logFileName.WriteString(".log")

	return logFileName.String()
}

func getLogFileWriter(filename string) lumberjack.Logger {
	return lumberjack.Logger{
		Filename:   filename, // 日志文件路径
		MaxSize:    10,       // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: 20,       // 日志文件最多保存多少个备份
		MaxAge:     7,        // 文件最多保存多少天
		Compress:   true,     // 是否压缩
	}
}

func getLevelLogFileWriter(serverName string, level string, path string) *lumberjack.Logger {
	fileWrite := getLogFileWriter(getLogFileName(serverName, level, path))
	return &fileWrite
}

func getLogCore(serverName string, stdConf *Conf, fileConf *Conf, reportConf *Conf) zapcore.Core {

	// 定义日志级别
	debugLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level == zapcore.DebugLevel
	})

	infoLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level == zapcore.InfoLevel
	})

	warnLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level == zapcore.WarnLevel
	})

	errorLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level >= zapcore.ErrorLevel
	})

	var zapCoreTee []zapcore.Core

	// 分别设置写入文件的参数、上报参数和显示到终端的参数
	if stdConf != nil {
		// 创建终端输出的目标 writer
		stdWriter := zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout))

		switch stdConf.Level {
		case Level.Debug:
			zapCoreTee = append(zapCoreTee, zapcore.NewCore(getStdEncoder(), stdWriter, debugLevel))
			fallthrough
		case Level.Info:
			zapCoreTee = append(zapCoreTee, zapcore.NewCore(getStdEncoder(), stdWriter, infoLevel))
			fallthrough
		case Level.Warn:
			zapCoreTee = append(zapCoreTee, zapcore.NewCore(getStdEncoder(), stdWriter, warnLevel))
			fallthrough
		case Level.Error:
			fallthrough
		default:
			zapCoreTee = append(zapCoreTee, zapcore.NewCore(getStdEncoder(), stdWriter, errorLevel))
		}
	}

	if fileConf != nil && fileConf.Path != "" {
		// 创建写入的目标 writer
		debugFileWriter := zapcore.NewMultiWriteSyncer(zapcore.AddSync(getLevelLogFileWriter(serverName, Level.Debug, fileConf.Path)))
		infoFileWriter := zapcore.NewMultiWriteSyncer(zapcore.AddSync(getLevelLogFileWriter(serverName, Level.Info, fileConf.Path)))
		warnFileWriter := zapcore.NewMultiWriteSyncer(zapcore.AddSync(getLevelLogFileWriter(serverName, Level.Warn, fileConf.Path)))
		errFileWriter := zapcore.NewMultiWriteSyncer(zapcore.AddSync(getLevelLogFileWriter(serverName, Level.Error, fileConf.Path)))

		switch fileConf.Level {
		case Level.Debug:
			zapCoreTee = append(zapCoreTee, zapcore.NewCore(getFileEncoder(), debugFileWriter, debugLevel))
			fallthrough
		case Level.Info:
			zapCoreTee = append(zapCoreTee, zapcore.NewCore(getFileEncoder(), infoFileWriter, infoLevel))
			fallthrough
		case Level.Warn:
			zapCoreTee = append(zapCoreTee, zapcore.NewCore(getFileEncoder(), warnFileWriter, warnLevel))
			fallthrough
		case Level.Error:
			fallthrough
		default:
			zapCoreTee = append(zapCoreTee, zapcore.NewCore(getFileEncoder(), errFileWriter, errorLevel))
		}
	}

	if reportConf != nil {
		if reportConf.CallBack == nil {
			reportConf.CallBack = func(level string, message string) {}
		}

		switch reportConf.Level {
		case Level.Debug:
			zapCoreTee = append(zapCoreTee, zapcore.NewCore(getReportEncoder(), zapcore.NewMultiWriteSyncer(zapcore.AddSync(report.IoWrite(Level.Debug, reportConf.CallBack))), debugLevel))
			fallthrough
		case Level.Info:
			zapCoreTee = append(zapCoreTee, zapcore.NewCore(getReportEncoder(), zapcore.NewMultiWriteSyncer(zapcore.AddSync(report.IoWrite(Level.Info, reportConf.CallBack))), infoLevel))
			fallthrough
		case Level.Warn:
			zapCoreTee = append(zapCoreTee, zapcore.NewCore(getReportEncoder(), zapcore.NewMultiWriteSyncer(zapcore.AddSync(report.IoWrite(Level.Warn, reportConf.CallBack))), warnLevel))
			fallthrough
		case Level.Error:
			fallthrough
		default:
			zapCoreTee = append(zapCoreTee, zapcore.NewCore(getReportEncoder(), zapcore.NewMultiWriteSyncer(zapcore.AddSync(report.IoWrite(Level.Error, reportConf.CallBack))), errorLevel))
		}
	}

	return zapcore.NewTee(zapCoreTee...)
}
