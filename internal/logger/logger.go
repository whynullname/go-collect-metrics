package logger

import (
	"go.uber.org/zap"
)

var Log *zap.SugaredLogger

func Initialize(level string) error {
	if Log != nil {
		return nil
	}

	lvl, err := zap.ParseAtomicLevel(level)

	if err != nil {
		return err
	}

	cfg := zap.NewDevelopmentConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	defer zl.Sync()

	if err != nil {
		return err
	}

	Log = zl.Sugar()
	return nil
}
