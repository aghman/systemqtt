package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// LoadOptions controls how configuration is merged (defaults, file, env, flags).
type LoadOptions struct {
	// ConfigFile is the resolved config path (--config or SYSTEMQTT_CONFIG_FILE). Empty skips file.
	ConfigFile string
	// Flags, if non-nil, is bound so CLI overrides env and file after parse.
	Flags *pflag.FlagSet
	// SkipValidate when true returns merged config without mqtt required fields (for config print/template).
	SkipValidate bool
}

// ResolveConfigFile returns --config if set, otherwise SYSTEMQTT_CONFIG_FILE.
func ResolveConfigFile(fs *pflag.FlagSet) string {
	if fs != nil {
		if f := fs.Lookup("config"); f != nil {
			if s := strings.TrimSpace(f.Value.String()); s != "" {
				return s
			}
		}
	}
	return strings.TrimSpace(os.Getenv("SYSTEMQTT_CONFIG_FILE"))
}

// Load merges defaults, optional config file, environment, and flags into Config.
// Precedence: defaults < config file < environment < flags.
func Load(opts LoadOptions) (*Config, error) {
	v := viper.New()
	setDefaults(v)

	if opts.ConfigFile != "" {
		v.SetConfigFile(opts.ConfigFile)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("read config file: %w", err)
		}
	}

	v.SetEnvPrefix("SYSTEMQTT")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if opts.Flags != nil {
		if err := bindPFlags(v, opts.Flags); err != nil {
			return nil, fmt.Errorf("bind flags: %w", err)
		}
	}

	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// Normalizations
	c.Log.Level = strings.ToLower(strings.TrimSpace(c.Log.Level))
	if c.Hostname == "" {
		h, err := os.Hostname()
		if err != nil {
			return nil, fmt.Errorf("hostname: %w", err)
		}
		c.Hostname = h
	}
	if strings.TrimSpace(c.MQTT.Client.ID) == "" {
		c.MQTT.Client.ID = "systemqtt-" + c.Hostname
	}

	if opts.SkipValidate {
		return &c, nil
	}
	return &c, c.Validate()
}

func bindPFlags(v *viper.Viper, fs *pflag.FlagSet) error {
	var err error
	fs.VisitAll(func(f *pflag.Flag) {
		if err != nil {
			return
		}
		key := f.Name
		if key == "config" {
			key = "config.file"
		}
		if e := v.BindPFlag(key, f); e != nil {
			err = e
		}
	})
	return err
}

// SetDefaultsForTemplate applies the same defaults as Load for template output.
func SetDefaultsForTemplate(v *viper.Viper) {
	setDefaults(v)
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("agent", false)
	v.SetDefault("config.file", "")
	v.SetDefault("mqtt.broker.url", "")
	v.SetDefault("mqtt.broker.port", 1883)
	v.SetDefault("mqtt.broker.username", "")
	v.SetDefault("mqtt.broker.password", "")
	v.SetDefault("mqtt.client.id", "")
	v.SetDefault("mqtt.qos", 1)
	v.SetDefault("mqtt.retain", true)
	v.SetDefault("mqtt.topic.prefix", "systemqtt")
	v.SetDefault("mqtt.tls.enabled", false)
	v.SetDefault("mqtt.tls.ca", "")
	v.SetDefault("mqtt.tls.cert", "")
	v.SetDefault("mqtt.tls.key", "")
	v.SetDefault("mqtt.tls.insecure", false)
	v.SetDefault("hostname", "")
	v.SetDefault("log.level", "info")
	v.SetDefault("systemd.unit.filter", "")
	v.SetDefault("systemd.publish.on_startup", true)
}
