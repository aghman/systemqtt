package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/systemqtt/systemqtt/internal/config"
)

// Run executes the root command and returns an exit code.
func Run() int {
	rootCmd := newRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "systemqtt",
		Short: "Publish systemd unit status to MQTT",
	}

	root.PersistentFlags().Bool("agent", false, "Format CLI output as JSON for agent consumption")
	config.RegisterPersistentFlags(root.PersistentFlags())

	root.AddCommand(newVersionCmd())
	root.AddCommand(newConfigCmd())
	root.AddCommand(newServeCmd())
	root.AddCommand(newDoctorCmd())

	return root
}
