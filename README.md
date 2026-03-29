# systemqtt

Go service that publishes **systemd unit state** to an **MQTT** broker over D-Bus (no polling). Retained JSON per unit, LWT offline string on the status topic, and optional TLS.

## Build

```bash
make build   # output: bin/systemqtt
make test
```

## Configuration

- **File**: YAML or JSON path via `--config` or `SYSTEMQTT_CONFIG_FILE`.
- **Environment**: `SYSTEMQTT_*` with nested keys as underscores (for example `SYSTEMQTT_MQTT_BROKER_URL`).
- **Flags**: same keys as the [design sketch](design/sketch.md) (for example `--mqtt.broker.url`).

Required: `mqtt.broker.url`. Username and password are optional for brokers that allow anonymous connections.

```bash
systemqtt config template
systemqtt config validate --config /etc/systemqtt/config.yaml
systemqtt doctor --config /etc/systemqtt/config.yaml
```

## MQTT

- Topics: `<prefix>/<hostname>/unit/<unit-name>/state` (JSON payload), `<prefix>/<hostname>/status` (`online` / `offline`).
- Defaults: prefix `systemqtt`, QoS `1`, retain `true`.
- Broker must be reachable from the host; use TLS options for production.

## Security

- Treat broker credentials as secrets: restrict file permissions on config files, prefer environment variables or a secret store for passwords.
- MQTT traffic should use TLS (`mqtt.tls.enabled` and CA/client certs) on untrusted networks.

## systemd

Example unit: [`deploy/systemqtt.service`](deploy/systemqtt.service). Adjust `ExecStart`, user, and groups so the process can reach the **system** D-Bus (typically root or a dedicated user with appropriate permissions). `WatchdogSec` expects periodic `WATCHDOG=1` notifications; the service sends them when systemd enables the watchdog.

```bash
sudo install -m755 bin/systemqtt /usr/local/bin/systemqtt
sudo install -m644 deploy/systemqtt.service /etc/systemd/system/systemqtt.service
sudo systemctl daemon-reload
sudo systemctl enable --now systemqtt
```

## Agent / JSON CLI

With `--agent`, commands emit a JSON envelope `{status,data,error}`; `serve` uses JSON logs in agent mode.
