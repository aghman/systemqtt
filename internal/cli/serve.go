package cli

import (
	"context"
	"net"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/systemqtt/systemqtt/internal/config"
	"github.com/systemqtt/systemqtt/internal/logger"
	mqttpub "github.com/systemqtt/systemqtt/internal/mqtt"
	"github.com/systemqtt/systemqtt/internal/systemd"
)

func newServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Connect to MQTT and publish systemd unit changes",
		RunE:  runServe,
	}
	return cmd
}

func runServe(cmd *cobra.Command, _ []string) error {
	agentMode, _ := cmd.Flags().GetBool("agent")
	cfgPath := config.ResolveConfigFile(cmd.Flags())
	cfg, err := config.Load(config.LoadOptions{ConfigFile: cfgPath, Flags: cmd.Flags()})
	if err != nil {
		return err
	}
	log, err := logger.New(cfg, agentMode)
	if err != nil {
		return err
	}
	defer func() { _ = log.Sync() }()

	pub, err := mqttpub.NewPublisher(cfg)
	if err != nil {
		return err
	}
	refreshAll := func() {
		units, err := systemd.ListMatchingUnits(cfg)
		if err != nil {
			log.Warn("full refresh failed", zap.Error(err))
			return
		}
		for _, u := range units {
			publishUnit(log, pub, cfg, u)
		}
	}
	pub.SetOnReconnect(refreshAll)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := pub.Connect(ctx); err != nil {
		return err
	}
	log.Info("connected to MQTT broker; online and publishing unit state",
		serveBrokerEndpoint(cfg))

	if cfg.Systemd.Publish.OnStartup {
		log.Info("loading initial systemd unit status for matching units")
		refreshAll()
	}

	w, err := systemd.NewWatcher(cfg)
	if err != nil {
		return err
	}
	defer func() { _ = w.Close() }()
	log.Info("connected to system D-Bus; subscribing to systemd unit status changes")

	go func() {
		if err := w.Run(ctx, func(ev systemd.UnitEvent) {
			publishUnit(log, pub, cfg, ev)
		}); err != nil && err != context.Canceled {
			log.Error("watcher exited", zap.Error(err))
		}
	}()

	notifyReady()
	startWatchdogIfEnabled(ctx, log)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	cancel()
	_ = pub.PublishOffline()
	pub.Disconnect()
	log.Info("shutdown complete")
	return nil
}

// serveBrokerEndpoint returns a zap field with host:port only (no credentials from URL).
func serveBrokerEndpoint(cfg *config.Config) zap.Field {
	raw := strings.TrimSpace(cfg.MQTT.Broker.URL)
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" {
		return zap.String("broker", "configured")
	}
	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = strconv.Itoa(cfg.MQTT.Broker.Port)
	}
	return zap.String("broker", net.JoinHostPort(host, port))
}

func publishUnit(log *zap.Logger, pub *mqttpub.Publisher, cfg *config.Config, ev systemd.UnitEvent) {
	st := mqttpub.UnitState{
		Unit:        ev.Unit,
		ActiveState: ev.ActiveState,
		SubState:    ev.SubState,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Hostname:    cfg.Hostname,
	}
	if err := pub.PublishUnitState(st); err != nil {
		log.Warn("publish unit failed", zap.String("unit", ev.Unit), zap.Error(err))
	}
}
