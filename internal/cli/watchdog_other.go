//go:build !linux

package cli

import (
	"context"

	"go.uber.org/zap"
)

func notifyReady() {}

func startWatchdogIfEnabled(_ context.Context, _ *zap.Logger) {}
