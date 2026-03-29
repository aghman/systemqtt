//go:build linux

package cli

import (
	"context"
	"time"

	"github.com/coreos/go-systemd/v22/daemon"
	"go.uber.org/zap"
)

func notifyReady() {
	_, _ = daemon.SdNotify(false, daemon.SdNotifyReady)
}

func startWatchdogIfEnabled(ctx context.Context, log *zap.Logger) {
	interval, err := daemon.SdWatchdogEnabled(false)
	if err != nil || interval <= 0 {
		return
	}
	t := time.NewTicker(interval / 3)
	go func() {
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				if _, err := daemon.SdNotify(false, daemon.SdNotifyWatchdog); err != nil {
					log.Debug("sd_notify watchdog", zap.Error(err))
				}
			}
		}
	}()
}
