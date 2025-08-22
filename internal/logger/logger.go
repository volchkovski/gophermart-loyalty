package logger

import (
	"fmt"
	"go.uber.org/zap"
)

const (
	LocalEnv      = "local"
	ProductionEnv = "prod"
)

var Log = zap.NewNop().Sugar()

func Initialize(level, env string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return fmt.Errorf("failed to parse log level: %w", err)
	}

	var cfg zap.Config
	switch env {
	case LocalEnv:
		cfg = zap.NewDevelopmentConfig()
	case ProductionEnv:
		cfg = zap.NewProductionConfig()
	default:
		return fmt.Errorf("logger valid environment values: %s, %s", LocalEnv, ProductionEnv)
	}

	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil {
		return fmt.Errorf("logger build fail: %w", err)
	}
	Log = zl.Sugar()
	return nil
}
