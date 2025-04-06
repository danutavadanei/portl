package common

import (
	"github.com/danutavadanei/portl/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(cfg *config.Config) (*zap.Logger, error) {
	logLevel := zapcore.InfoLevel
	if cfg.Debug {
		logLevel = zapcore.DebugLevel
	}

	zapCfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(logLevel),
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := zapCfg.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}
