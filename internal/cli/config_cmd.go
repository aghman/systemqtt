package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/systemqtt/systemqtt/internal/agent"
	"github.com/systemqtt/systemqtt/internal/config"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Inspect or validate configuration",
	}
	cmd.AddCommand(newConfigPrintCmd())
	cmd.AddCommand(newConfigTemplateCmd())
	cmd.AddCommand(newConfigValidateCmd())
	return cmd
}

func newConfigPrintCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "print",
		Short: "Print merged configuration (defaults, file, env, flags)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			agentMode, _ := cmd.Flags().GetBool("agent")
			cfgPath := config.ResolveConfigFile(cmd.Flags())
			cfg, err := config.Load(config.LoadOptions{
				ConfigFile:   cfgPath,
				Flags:        cmd.Flags(),
				SkipValidate: true,
			})
			if err != nil {
				if agentMode {
					_ = agent.WriteError(cmd.OutOrStdout(), err.Error())
				} else {
					fmt.Fprintln(cmd.ErrOrStderr(), err)
				}
				return err
			}
			red := cfg.Redacted()
			if agentMode {
				return agent.WriteOK(cmd.OutOrStdout(), red)
			}
			m, err := configAsYAMLMap(red)
			if err != nil {
				return err
			}
			enc := yaml.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent(2)
			if err := enc.Encode(m); err != nil {
				return err
			}
			return enc.Close()
		},
	}
	cmd.SilenceErrors = true
	return cmd
}

func newConfigTemplateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "template",
		Short: "Print a configuration file template with defaults",
		RunE: func(cmd *cobra.Command, _ []string) error {
			agentMode, _ := cmd.Flags().GetBool("agent")
			v := viper.New()
			config.SetDefaultsForTemplate(v)
			all := v.AllSettings()
			if agentMode {
				raw, err := json.Marshal(all)
				if err != nil {
					return err
				}
				return agent.WriteOK(cmd.OutOrStdout(), json.RawMessage(raw))
			}
			enc := yaml.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent(2)
			if err := enc.Encode(all); err != nil {
				return err
			}
			return enc.Close()
		},
	}
}

func newConfigValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate merged configuration",
		RunE: func(cmd *cobra.Command, _ []string) error {
			agentMode, _ := cmd.Flags().GetBool("agent")
			cfgPath := config.ResolveConfigFile(cmd.Flags())
			_, err := config.Load(config.LoadOptions{ConfigFile: cfgPath, Flags: cmd.Flags()})
			if err != nil {
				if agentMode {
					_ = agent.WriteError(cmd.OutOrStdout(), err.Error())
				} else {
					fmt.Fprintln(cmd.ErrOrStderr(), err)
				}
				return err
			}
			if agentMode {
				return agent.WriteOK(cmd.OutOrStdout(), map[string]string{"result": "valid"})
			}
			fmt.Fprintln(cmd.OutOrStdout(), "configuration is valid")
			return nil
		},
	}
	cmd.SilenceErrors = true
	return cmd
}

func configAsYAMLMap(c config.Config) (map[string]interface{}, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}
