package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
)

func TestValidate_requiredBrokerURL(t *testing.T) {
	c := minimalValidConfig()
	c.MQTT.Broker.URL = ""
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for empty broker URL")
	}
}

func TestValidate_anonymousMQTTCredentials(t *testing.T) {
	c := minimalValidConfig()
	c.MQTT.Broker.Username = ""
	c.MQTT.Broker.Password = ""
	if err := c.Validate(); err != nil {
		t.Fatalf("expected valid config with empty broker credentials: %v", err)
	}
}

func TestValidate_qos(t *testing.T) {
	c := minimalValidConfig()
	c.MQTT.QoS = 3
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for invalid qos")
	}
}

func TestValidate_logLevel(t *testing.T) {
	c := minimalValidConfig()
	c.Log.Level = "nope"
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for invalid log level")
	}
}

func TestValidate_glob(t *testing.T) {
	c := minimalValidConfig()
	c.Systemd.Unit.Filter = "["
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for invalid glob")
	}
}

func TestLoad_fromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.yaml")
	content := `
mqtt:
  broker:
    url: "tcp://broker.example.com"
    port: 8883
    username: "u"
    password: "p"
log:
  level: warn
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(LoadOptions{ConfigFile: path})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MQTT.Broker.Port != 8883 {
		t.Fatalf("port: got %d", cfg.MQTT.Broker.Port)
	}
	if cfg.Log.Level != "warn" {
		t.Fatalf("log level: got %q", cfg.Log.Level)
	}
}

func TestLoad_flagOverridesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.yaml")
	if err := os.WriteFile(path, []byte(`
mqtt:
  broker:
    url: "tcp://a.example"
    username: "u"
    password: "p"
`), 0o600); err != nil {
		t.Fatal(err)
	}
	fs := pflag.NewFlagSet("t", pflag.ContinueOnError)
	RegisterFlags(fs)
	if err := fs.Set("mqtt.broker.url", "tcp://b.example"); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(LoadOptions{ConfigFile: path, Flags: fs})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MQTT.Broker.URL != "tcp://b.example" {
		t.Fatalf("got %q", cfg.MQTT.Broker.URL)
	}
}

func minimalValidConfig() *Config {
	c := &Config{}
	c.MQTT.Broker.URL = "tcp://localhost"
	c.MQTT.Broker.Port = 1883
	c.MQTT.QoS = 1
	c.Log.Level = "info"
	return c
}
