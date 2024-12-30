package logger

import (
	"os"

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

	terminalSyncer := zapcore.AddSync(os.Stdout)

	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())

	fileCore := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)
	terminalCore := zapcore.NewCore(encoder, terminalSyncer, zapcore.DebugLevel)

	logger := zap.New(zapcore.NewTee(fileCore, terminalCore))

	return logger
}
