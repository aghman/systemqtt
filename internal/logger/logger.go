package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/systemqtt/systemqtt/internal/config"
)

// New returns a zap logger for serve: JSON when agent mode, console otherwise.
func New(cfg *config.Config, agent bool) (*zap.Logger, error) {
	lv, err := zapcore.ParseLevel(strings.ToLower(cfg.Log.Level))
	if err != nil {
		return nil, err
	}
	if agent {
		enc := zap.NewProductionEncoderConfig()
		enc.EncodeTime = zapcore.ISO8601TimeEncoder
		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(enc),
			zapcore.AddSync(os.Stdout),
			zap.NewAtomicLevelAt(lv),
		)
		return zap.New(core), nil
	}
	enc := zap.NewDevelopmentEncoderConfig()
	enc.EncodeLevel = zapcore.CapitalColorLevelEncoder
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(enc),
		zapcore.AddSync(os.Stdout),
		zap.NewAtomicLevelAt(lv),
	)
	return zap.New(core), nil
}
