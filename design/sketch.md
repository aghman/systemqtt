# Sketch

## Overview

The goal of this service is to provide a way to publish systemd unit status to MQTT.
It will be a simple service that will listen for systemd unit status changes and publish them to MQTT.

## Design

### Architecture

The service has three core components:

1. **Systemd Watcher** — Subscribes to `org.freedesktop.systemd1` property changes over D-Bus using `github.com/godbus/dbus/v5`. Event-driven; no polling. On startup, performs an initial scan to publish current state of all matching units so subscribers get a full picture immediately.

2. **MQTT Publisher** — Maintains a persistent connection to the MQTT broker using `eclipse/paho.mqtt.golang`. Publishes unit state changes as retained JSON messages at QoS 1. Registers a Last Will and Testament (LWT) so the broker publishes an offline status if the service disconnects unexpectedly. Auto-reconnects with exponential backoff on connection loss; publishes a full state refresh on reconnect.

3. **Config / CLI Layer** — Cobra commands backed by Viper for unified config from CLI flags, environment variables, and config files. Zap for structured logging.

### MQTT Topic Structure

```
<prefix>/<hostname>/unit/<unit-name>/state      — full JSON state payload
<prefix>/<hostname>/status                      — "online" / LWT "offline"
```

- `<prefix>` defaults to `systemqtt`, configurable via `mqtt.topic.prefix`.
- `<hostname>` defaults to `os.Hostname()`, configurable via `hostname`.
- `<unit-name>` is the full systemd unit name (e.g. `nginx.service`).

### MQTT Message Payload

Unit state messages are JSON:

```json
{
  "unit": "nginx.service",
  "active_state": "active",
  "sub_state": "running",
  "timestamp": "2026-03-29T12:00:00Z",
  "hostname": "web-01"
}
```

The `<prefix>/<hostname>/status` topic carries a simple string: `"online"` on connect, `"offline"` via LWT.

### Systemd Unit Filtering

The `systemd.unit.filter` option accepts a comma-separated list of glob patterns matched against unit names. Examples:

- `nginx.service` — exact match
- `*.service` — all service units
- `myapp-*` — prefix match

If not set, all units are published.

### Implementation

- Uses `github.com/godbus/dbus/v5` for D-Bus subscription
- Uses `eclipse/paho.mqtt.golang` for MQTT
- Uses `spf13/cobra` for command line parsing
- Uses `spf13/viper` for configuration
- Uses `uber-go/zap` for logging
- Uses Makefiles for building

### Configuration

- The service can be configured via a configuration file in JSON or YAML format. Viper auto-detects format from the file extension.
- The configuration file location can be passed as a command line parameter or via environment variable.
- All configuration options are available as environment variables, configuration file options, and command line parameters.
- All configuration options are validated at startup; the service exits with a clear error if the configuration is invalid.
- All configuration options have a default value where a sensible default exists; options without a sensible default (broker URL) are required and validated for presence. MQTT username and password default to empty for brokers that allow anonymous connections.
- All configuration options have a description.

### Configuration Options

| Option | Description | Default | Required |
|---|---|---|---|
| `agent` | Format CLI output as JSON for AI agent consumption | `false` | No |
| `config.file` | Path to the configuration file | `""` | No |
| `mqtt.broker.url` | Broker URL with scheme and host (e.g. `tcp://mqtt.example.com`); not a bare hostname. Port optional; if omitted, `mqtt.broker.port` is used | — | Yes |
| `mqtt.broker.port` | Port of the MQTT broker | `1883` | No |
| `mqtt.broker.username` | Username for MQTT broker authentication (omit for anonymous brokers) | `""` | No |
| `mqtt.broker.password` | Password for MQTT broker authentication (omit for anonymous brokers) | `""` | No |
| `mqtt.client.id` | MQTT client ID | `systemqtt-<hostname>` | No |
| `mqtt.qos` | MQTT QoS level (0, 1, 2) | `1` | No |
| `mqtt.retain` | Retain MQTT messages | `true` | No |
| `mqtt.tls.enabled` | Enable TLS for MQTT connection | `false` | No |
| `mqtt.tls.ca` | Path to CA certificate | `""` | No |
| `mqtt.tls.cert` | Path to client certificate | `""` | No |
| `mqtt.tls.key` | Path to client key | `""` | No |
| `mqtt.tls.insecure` | Skip TLS certificate verification | `false` | No |
| `mqtt.topic.prefix` | Base prefix for MQTT topics | `systemqtt` | No |
| `hostname` | Hostname used in MQTT topics | OS hostname | No |
| `log.level` | Log level: debug, info, warn, error, fatal | `info` | No |
| `systemd.unit.filter` | Comma-separated glob patterns for unit names to publish | `""` (all units) | No |
| `systemd.publish.on_startup` | Publish current state of all matching units on startup | `true` | No |

### Commands

- `config` — Prints the current resolved configuration (all sources merged) and their values.
- `config template` — Outputs a template configuration file with all options and their default values.
- `config validate` — Validates the configuration, prints success or failure, and identifies any invalid or missing required options.
- `doctor` — Validates configuration from all sources, attempts to connect to the MQTT broker, and verifies D-Bus access to systemd. Prints a summary of pass/fail checks.
- `serve` — Starts the service: connects to MQTT, subscribes to D-Bus, publishes unit state changes.
- `version` — Prints the version of the service.

### Agent Mode

When `--agent` is passed, all command output is structured JSON with a consistent envelope:

```json
{
  "status": "ok",
  "data": { ... },
  "error": null
}
```

On failure:

```json
{
  "status": "error",
  "data": null,
  "error": "mqtt broker connection refused"
}
```

For the `serve` command, agent mode switches zap to JSON-encoded structured logging.

## Runtime Considerations

- The service runs as a systemd unit. A `.service` file will be provided.
- The service should be run as a user that has access to D-Bus and systemd unit status.
- Graceful shutdown on `SIGTERM` and `SIGINT`: explicitly publishes the offline status to the LWT topic, unsubscribes from D-Bus, disconnects from MQTT, and exits with code 0.
- Supports systemd `WatchdogSec=` — periodically notifies systemd the process is alive.
