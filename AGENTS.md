# systemqtt

A Go-based service that publishes systemd unit status to MQTT.

## Cursor Cloud specific instructions

### Services

| Service | Purpose | How to start |
|---|---|---|
| **Mosquitto** (MQTT broker) | Required for the app to publish messages | `mosquitto -d -p 1883` |
| **systemd** | Provides unit status data | Available on host (limited in containers) |

### Development commands

- **Build:** `go build -o systemqtt .`
- **Run:** `go run .`
- **Test:** `go test -v ./...`
- **Lint:** `golangci-lint run ./...`
- **Tidy deps:** `go mod tidy`

### Environment variables

| Variable | Default | Description |
|---|---|---|
| `MQTT_BROKER` | `tcp://localhost:1883` | MQTT broker connection string |
| `SYSTEMD_UNITS` | `mosquitto` | Comma-separated list of systemd units to monitor |

### Gotchas

- `GOPATH/bin` must be on `PATH` for `golangci-lint` to work. It is configured in `~/.bashrc` but may not be sourced in non-interactive shells. Use `export PATH="$PATH:$(go env GOPATH)/bin"` if needed.
- In container/VM environments, `systemctl` may report units as inactive/unknown since systemd is not the init system. This is expected — the MQTT publish pipeline still works correctly.
- The MQTT library (`paho.mqtt.golang`) requires Go >= 1.24.0. The Go toolchain auto-upgrades via `go.mod` toolchain directive.
- Start Mosquitto before running the app: `mosquitto -d -p 1883`. Use `mosquitto_pub`/`mosquitto_sub` CLI tools for manual testing.
