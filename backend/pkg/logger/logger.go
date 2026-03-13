package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var globalLogger *zap.Logger

// Init 初始化日志
func Init(level, logFile string, maxSize, maxBackups, maxAge int) error {
	// 创建日志目录
	if logFile != "" {
		dir := filepath.Dir(logFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// 解析日志级别
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// 编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 控制台输出
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	consoleCore := zapcore.NewCore(
		consoleEncoder,
		zapcore.AddSync(os.Stdout),
		zapLevel,
	)
	// 文件输出
	var cores []zapcore.Core
	cores = append(cores, consoleCore)

	if logFile != "" {
		fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
		fileWriter := zapcore.AddSync(&zapcore.BufferedWriteSyncer{
			WS: zapcore.AddSync(&fileWriter{path: logFile}),
		})
		fileCore := zapcore.NewCore(fileEncoder, fileWriter, zapLevel)
		cores = append(cores, fileCore)
	}

	// 创建 logger
	core := zapcore.NewTee(cores...)
	globalLogger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return nil
}

// Get 获取全局 logger
func Get() *zap.Logger {
	if globalLogger == nil {
		// 如果未初始化，使用默认配置
		globalLogger, _ = zap.NewProduction()
	}
	return globalLogger
}

// Debug 调试日志
func Debug(msg string, fields ...zap.Field) {
	Get().Debug(msg, fields...)
}

// Info 信息日志
func Info(msg string, fields ...zap.Field) {
	Get().Info(msg, fields...)
}

// Warn 警告日志
func Warn(msg string, fields ...zap.Field) {
	Get().Warn(msg, fields...)
}

// Error 错误日志
func Error(msg string, fields ...zap.Field) {
	Get().Error(msg, fields...)
}

// Fatal 致命错误日志
func Fatal(msg string, fields ...zap.Field) {
	Get().Fatal(msg, fields...)
}

// Sync 同步日志
func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}

// fileWriter 简单的文件写入器
type fileWriter struct {
	path string
	file *os.File
}

func (w *fileWriter) Write(p []byte) (n int, err error) {
	if w.file == nil {
		w.file, err = os.OpenFile(w.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return 0, err
		}
	}
	return w.file.Write(p)
}

func (w *fileWriter) Sync() error {
	if w.file != nil {
		return w.file.Sync()
	}
	return nil
}
