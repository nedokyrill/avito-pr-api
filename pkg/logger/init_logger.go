package logger

import (
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.SugaredLogger

func InitLogger() {

	dev := false
	encoderCfg := zap.NewProductionEncoderConfig()
	level := zap.NewAtomicLevelAt(zap.InfoLevel)

	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	config := zap.Config{
		Level:             level,
		Development:       dev,
		DisableStacktrace: true,
		DisableCaller:     false,
		Sampling:          nil,
		Encoding:          "json",
		EncoderConfig:     encoderCfg,
		OutputPaths: []string{
			"logs/log.txt",
			"stdout",
		},
		ErrorOutputPaths: []string{
			"logs/error.txt",
			"stderr",
		},
		InitialFields: map[string]interface{}{
			"pid": os.Getpid(),
		},
	}

	baseLogger, err := config.Build()
	if err != nil {
		log.Fatal("Error building zap baseLogger")
	}

	Logger = baseLogger.Sugar()
}
