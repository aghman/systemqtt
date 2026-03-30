# systemqtt Configuration Guide

This guide explains how to configure `systemqtt` with:

- configuration files (YAML or JSON)
- environment variables (`SYSTEMQTT_*`)
- command-line flags (`--mqtt.*`, `--systemd.*`, etc.)

## Configuration Sources and Precedence

`systemqtt` merges configuration in this order (lowest to highest priority):

1. built-in defaults
2. config file (`--config` or `SYSTEMQTT_CONFIG_FILE`)
3. environment variables (`SYSTEMQTT_*`)
4. command-line flags

If the same setting appears in multiple places, the source with higher priority wins.

## Quick Start

Use these commands to generate, inspect, and validate configuration:

```bash
# Print a template config with defaults
systemqtt config template

# Validate a config file (plus env/flags if set)
systemqtt config validate --config /etc/systemqtt/config.yaml

# Print final merged config (with password redacted)
systemqtt config print --config /etc/systemqtt/config.yaml

# Run health checks (config + MQTT + systemd access)
systemqtt doctor --config /etc/systemqtt/config.yaml
```

## Option Reference

All keys can be set through file/env/flags.

| Key | Type | Default | Required | Notes |
|---|---|---|---|---|
| `agent` | bool | `false` | No | JSON envelope output for CLI commands |
| `config.file` | string | `""` | No | Usually set by `--config` or `SYSTEMQTT_CONFIG_FILE` |
| `mqtt.broker.url` | string | `""` | Yes | Must include scheme and host, for example `tcp://mqtt.example.com` |
| `mqtt.broker.port` | int | `1883` | No | Range `1..65535` |
| `mqtt.broker.username` | string | `""` | No | Optional for anonymous brokers |
| `mqtt.broker.password` | string | `""` | No | Secret value; redacted by `config print` |
| `mqtt.client.id` | string | `systemqtt-<hostname>` | No | Auto-derived when empty |
| `mqtt.qos` | int | `1` | No | Must be `0`, `1`, or `2` |
| `mqtt.retain` | bool | `true` | No | Retained publishes for unit state |
| `mqtt.topic.prefix` | string | `systemqtt` | No | Topic root prefix |
| `mqtt.tls.enabled` | bool | `false` | No | Enables TLS settings below |
| `mqtt.tls.ca` | string | `""` | No | Optional CA cert path |
| `mqtt.tls.cert` | string | `""` | No | Client cert path (typically paired with key) |
| `mqtt.tls.key` | string | `""` | No | Client private key path |
| `mqtt.tls.insecure` | bool | `false` | No | Skip cert verification (not recommended for production) |
| `hostname` | string | OS hostname | No | Used in MQTT topic path |
| `log.level` | string | `info` | No | One of `debug`, `info`, `warn`, `error`, `fatal` |
| `systemd.unit.filter` | string | `""` | No | Comma-separated glob list, for example `nginx.service,*.timer` |
| `systemd.publish.on_startup` | bool | `true` | No | Publish current state at startup |

## File-Based Configuration

`systemqtt` accepts YAML or JSON files (detected by extension).

### YAML example (`/etc/systemqtt/config.yaml`)

```yaml
mqtt:
  broker:
    url: "tcp://mqtt.lan"
    port: 1883
    username: "systemqtt"
    password: "replace-me"
  client:
    id: "systemqtt-web-01"
  qos: 1
  retain: true
  topic:
    prefix: "systemqtt"
  tls:
    enabled: false
    ca: ""
    cert: ""
    key: ""
    insecure: false

hostname: "web-01"
log:
  level: "info"
systemd:
  unit:
    filter: "*.service,*.timer"
  publish:
    on_startup: true
```

### JSON example (`/etc/systemqtt/config.json`)

```json
{
  "mqtt": {
    "broker": {
      "url": "tcp://mqtt.example.net",
      "port": 1883,
      "username": "",
      "password": ""
    },
    "client": {
      "id": ""
    },
    "qos": 1,
    "retain": true,
    "topic": {
      "prefix": "systemqtt"
    },
    "tls": {
      "enabled": true,
      "ca": "/etc/ssl/certs/ca-certificates.crt",
      "cert": "/etc/systemqtt/client.crt",
      "key": "/etc/systemqtt/client.key",
      "insecure": false
    }
  },
  "hostname": "edge-01",
  "log": {
    "level": "info"
  },
  "systemd": {
    "unit": {
      "filter": "ssh.service,docker.service"
    },
    "publish": {
      "on_startup": true
    }
  }
}
```

## Environment Variables

Environment variables use:

- prefix: `SYSTEMQTT_`
- key mapping: dots become underscores
- uppercase keys

Examples:

- `mqtt.broker.url` -> `SYSTEMQTT_MQTT_BROKER_URL`
- `systemd.publish.on_startup` -> `SYSTEMQTT_SYSTEMD_PUBLISH_ON_STARTUP`
- `log.level` -> `SYSTEMQTT_LOG_LEVEL`

### Env-only example

```bash
export SYSTEMQTT_MQTT_BROKER_URL="tcp://mqtt.local"
export SYSTEMQTT_MQTT_BROKER_PORT="1883"
export SYSTEMQTT_MQTT_TOPIC_PREFIX="infra"
export SYSTEMQTT_SYSTEMD_UNIT_FILTER="nginx.service,sshd.service"
export SYSTEMQTT_LOG_LEVEL="debug"

systemqtt serve
```

### Config file path via env

```bash
export SYSTEMQTT_CONFIG_FILE="/etc/systemqtt/config.yaml"
systemqtt config validate
systemqtt serve
```

## Command-Line Flags

Flags use the same dotted keys as config options:

```bash
systemqtt serve \
  --mqtt.broker.url "tcp://mqtt.local" \
  --mqtt.broker.port 1883 \
  --mqtt.topic.prefix "lab" \
  --hostname "web-02" \
  --systemd.unit.filter "*.service" \
  --log.level "info"
```

For ad-hoc runs, flags are convenient. For services, prefer file + environment variables.

## Practical Override Pattern

A common production setup:

1. Keep non-secret defaults in `/etc/systemqtt/config.yaml`.
2. Inject secrets from environment (for example via systemd `Environment=` or `EnvironmentFile=`).
3. Use flags only for temporary overrides or debugging.

Example:

```bash
SYSTEMQTT_MQTT_BROKER_PASSWORD="$(cat /run/secrets/systemqtt_mqtt_password)" \
systemqtt serve --config /etc/systemqtt/config.yaml
```

## Validation Rules and Common Errors

Validation checks include:

- `mqtt.broker.url` is required and must parse as a URL
- `mqtt.broker.port` must be `1..65535`
- `mqtt.qos` must be `0`, `1`, or `2`
- `log.level` must be one of `debug|info|warn|error|fatal`
- `systemd.unit.filter` must be valid glob syntax
- when TLS is enabled, configured TLS file paths must exist

Typical errors:

- `mqtt.broker.url is required`
- `mqtt.qos must be 0, 1, or 2`
- `systemd.unit.filter: invalid glob pattern ...`
- `mqtt.tls.cert: stat /path/to/client.crt: no such file or directory`

## Recommended Deployment Layout

- Binary: `/usr/local/bin/systemqtt`
- Config: `/etc/systemqtt/config.yaml`
- Optional secret env file: `/etc/systemqtt/systemqtt.env` (`chmod 600`)
- Unit file: `/etc/systemd/system/systemqtt.service`

Then:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now systemqtt
sudo systemctl status systemqtt
```
