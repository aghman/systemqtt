package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/systemqtt/systemqtt/internal/agent"
	"github.com/systemqtt/systemqtt/internal/config"
	mqttpub "github.com/systemqtt/systemqtt/internal/mqtt"
	"github.com/systemqtt/systemqtt/internal/systemd"
)

type checkResult struct {
	Name  string `json:"name"`
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

func newDoctorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Validate configuration, MQTT, and systemd D-Bus access",
		RunE:  runDoctor,
	}
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	return cmd
}

func runDoctor(cmd *cobra.Command, _ []string) error {
	agentMode, _ := cmd.Flags().GetBool("agent")
	cfgPath := config.ResolveConfigFile(cmd.Flags())

	cfg, cerr := config.Load(config.LoadOptions{ConfigFile: cfgPath, Flags: cmd.Flags()})
	results := []checkResult{
		{Name: "config", OK: cerr == nil, Error: errString(cerr)},
	}

	if cfg != nil {
		pub, err := mqttpub.NewPublisher(cfg)
		if err != nil {
			results = append(results, checkResult{Name: "mqtt", OK: false, Error: err.Error()})
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			connErr := pub.Connect(ctx)
			cancel()
			pub.Disconnect()
			results = append(results, checkResult{Name: "mqtt", OK: connErr == nil, Error: errString(connErr)})
		}

		_, werr := systemd.NewWatcher(cfg)
		results = append(results, checkResult{Name: "dbus_systemd", OK: werr == nil, Error: errString(werr)})
	} else {
		results = append(results,
			checkResult{Name: "mqtt", OK: false, Error: "skipped (invalid config)"},
			checkResult{Name: "dbus_systemd", OK: false, Error: "skipped (invalid config)"},
		)
	}

	allOK := true
	for _, r := range results {
		if !r.OK {
			allOK = false
			break
		}
	}

	if agentMode {
		if allOK {
			if err := agent.WriteOK(cmd.OutOrStdout(), map[string]any{"checks": results}); err != nil {
				return err
			}
			return nil
		}
		_ = agent.WriteError(cmd.OutOrStdout(), "one or more checks failed")
		return fmt.Errorf("doctor checks failed")
	}

	for _, r := range results {
		status := "PASS"
		if !r.OK {
			status = "FAIL"
		}
		if r.Error != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s — %s\n", status, r.Name, r.Error)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "%s %s\n", status, r.Name)
		}
	}
	if !allOK {
		return fmt.Errorf("doctor checks failed")
	}
	fmt.Fprintln(cmd.OutOrStdout(), "All checks passed.")
	return nil
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
