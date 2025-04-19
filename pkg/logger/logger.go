package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func SetupLogger(env string) *zap.Logger {
	var logger *zap.Logger

	switch env {
	case envLocal:
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder        // Читаемая временная метка
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Цветные уровни логов
		config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder      // Длинный путь к файлу
		config.OutputPaths = []string{"stderr"}

		logger, _ = config.Build()
	case envProd:
		logger = zap.Must(zap.NewProduction())
	}

	return logger
}
