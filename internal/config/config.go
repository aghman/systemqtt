// Package config defines application settings, Viper loading, and validation.
package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Config holds all runtime options from sketch.md.
type Config struct {
	Agent bool `mapstructure:"agent"`
	// Config holds meta-options (e.g. config file path when embedded in merged view).
	Config struct {
		File string `mapstructure:"file"`
	} `mapstructure:"config"`
	MQTT struct {
		Broker struct {
			URL      string `mapstructure:"url"`
			Port     int    `mapstructure:"port"`
			Username string `mapstructure:"username"`
			Password string `mapstructure:"password"`
		} `mapstructure:"broker"`
		Client struct {
			ID string `mapstructure:"id"`
		} `mapstructure:"client"`
		QoS    int  `mapstructure:"qos"`
		Retain bool `mapstructure:"retain"`
		Topic  struct {
			Prefix string `mapstructure:"prefix"`
		} `mapstructure:"topic"`
		TLS struct {
			Enabled  bool   `mapstructure:"enabled"`
			CA       string `mapstructure:"ca"`
			Cert     string `mapstructure:"cert"`
			Key      string `mapstructure:"key"`
			Insecure bool   `mapstructure:"insecure"`
		} `mapstructure:"tls"`
	} `mapstructure:"mqtt"`
	Hostname string `mapstructure:"hostname"`
	Log      struct {
		Level string `mapstructure:"level"`
	} `mapstructure:"log"`
	Systemd struct {
		Unit struct {
			Filter string `mapstructure:"filter"`
		} `mapstructure:"unit"`
		Publish struct {
			OnStartup bool `mapstructure:"on_startup"`
		} `mapstructure:"publish"`
	} `mapstructure:"systemd"`
}

// Validate checks required fields and value ranges.
func (c *Config) Validate() error {
	var errs []error
	if strings.TrimSpace(c.MQTT.Broker.URL) == "" {
		errs = append(errs, errors.New("mqtt.broker.url is required"))
	} else if _, err := url.Parse(c.MQTT.Broker.URL); err != nil {
		errs = append(errs, fmt.Errorf("mqtt.broker.url: invalid URL: %w", err))
	}
	if c.MQTT.Broker.Port < 1 || c.MQTT.Broker.Port > 65535 {
		errs = append(errs, fmt.Errorf("mqtt.broker.port must be between 1 and 65535, got %d", c.MQTT.Broker.Port))
	}
	if strings.TrimSpace(c.MQTT.Broker.Username) == "" {
		errs = append(errs, errors.New("mqtt.broker.username is required"))
	}
	if c.MQTT.Broker.Password == "" {
		errs = append(errs, errors.New("mqtt.broker.password is required"))
	}
	if c.MQTT.QoS < 0 || c.MQTT.QoS > 2 {
		errs = append(errs, fmt.Errorf("mqtt.qos must be 0, 1, or 2, got %d", c.MQTT.QoS))
	}
	switch strings.ToLower(c.Log.Level) {
	case "debug", "info", "warn", "error", "fatal":
	default:
		errs = append(errs, fmt.Errorf("log.level must be one of debug, info, warn, error, fatal, got %q", c.Log.Level))
	}
	if err := validateGlobList(c.Systemd.Unit.Filter); err != nil {
		errs = append(errs, fmt.Errorf("systemd.unit.filter: %w", err))
	}
	if c.MQTT.TLS.Enabled {
		if strings.TrimSpace(c.MQTT.TLS.CA) != "" {
			if _, err := os.Stat(c.MQTT.TLS.CA); err != nil {
				errs = append(errs, fmt.Errorf("mqtt.tls.ca: %w", err))
			}
		}
		if strings.TrimSpace(c.MQTT.TLS.Cert) != "" || strings.TrimSpace(c.MQTT.TLS.Key) != "" {
			if _, err := os.Stat(c.MQTT.TLS.Cert); err != nil {
				errs = append(errs, fmt.Errorf("mqtt.tls.cert: %w", err))
			}
			if _, err := os.Stat(c.MQTT.TLS.Key); err != nil {
				errs = append(errs, fmt.Errorf("mqtt.tls.key: %w", err))
			}
		}
	}
	return joinErrs(errs)
}

func validateGlobList(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if _, err := filepath.Match(p, "placeholder"); err != nil {
			return fmt.Errorf("invalid glob pattern %q: %w", p, err)
		}
	}
	return nil
}

func joinErrs(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	var b strings.Builder
	for i, e := range errs {
		if i > 0 {
			b.WriteString("; ")
		}
		b.WriteString(e.Error())
	}
	return errors.New(b.String())
}

// Redacted returns a copy safe for logs (password masked).
func (c *Config) Redacted() Config {
	out := *c
	if out.MQTT.Broker.Password != "" {
		out.MQTT.Broker.Password = "[redacted]"
	}
	return out
}
