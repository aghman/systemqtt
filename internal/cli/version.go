package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/systemqtt/systemqtt/internal/agent"
	"github.com/systemqtt/systemqtt/internal/version"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(cmd *cobra.Command, _ []string) error {
			agentMode, _ := cmd.Flags().GetBool("agent")
			if agentMode {
				return agent.WriteOK(cmd.OutOrStdout(), map[string]string{
					"version":    version.Version,
					"git_commit": version.GitCommit,
					"build_date": version.BuildDate,
					"summary":    version.String(),
				})
			}
			fmt.Fprintln(cmd.OutOrStdout(), version.String())
			return nil
		},
	}
}
