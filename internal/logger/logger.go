package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func NewLogger(logFile string) *zap.Logger {
	writeSyncer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    10, // MB
		MaxBackups: 5,
		MaxAge:     30, // Days
		Compress:   true,
	})

	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())

	core := zapcore.NewCore(encoder, writeSyncer, zapcore.InfoLevel)
	logger := zap.New(core)

	return logger
}
