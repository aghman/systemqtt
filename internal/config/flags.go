package config

import (
	"github.com/spf13/pflag"
)

// RegisterPersistentFlags registers global flags on the root command (alias for RegisterFlags).
func RegisterPersistentFlags(fs *pflag.FlagSet) {
	RegisterFlags(fs)
}

// RegisterFlags defines CLI flags for all sketch options. Bind with Load(..., Flags: fs).
// The root command registers --agent separately; do not duplicate it here.
func RegisterFlags(fs *pflag.FlagSet) {
	fs.String("config", "", "Path to configuration file (YAML or JSON)")
	fs.String("mqtt.broker.url", "", "MQTT broker URL (required)")
	fs.Int("mqtt.broker.port", 1883, "MQTT broker port")
	fs.String("mqtt.broker.username", "", "MQTT username (omit for anonymous brokers)")
	fs.String("mqtt.broker.password", "", "MQTT password (omit for anonymous brokers)")
	fs.String("mqtt.client.id", "", "MQTT client ID (default systemqtt-<hostname>)")
	fs.Int("mqtt.qos", 1, "MQTT QoS (0-2)")
	fs.Bool("mqtt.retain", true, "MQTT retain flag")
	fs.String("mqtt.topic.prefix", "systemqtt", "Topic prefix for MQTT")
	fs.Bool("mqtt.tls.enabled", false, "Enable TLS for MQTT")
	fs.String("mqtt.tls.ca", "", "Path to CA certificate")
	fs.String("mqtt.tls.cert", "", "Path to client certificate")
	fs.String("mqtt.tls.key", "", "Path to client key")
	fs.Bool("mqtt.tls.insecure", false, "Skip TLS certificate verification")
	fs.String("hostname", "", "Hostname segment in MQTT topics")
	fs.String("log.level", "info", "Log level: debug, info, warn, error, fatal")
	fs.String("systemd.unit.filter", "", "Comma-separated glob patterns for units")
	fs.Bool("systemd.publish.on_startup", true, "Publish all matching units on startup")
}
