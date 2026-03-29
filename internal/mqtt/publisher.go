package mqtt

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/systemqtt/systemqtt/internal/config"
)

// UnitState is the JSON payload for unit state topics.
type UnitState struct {
	Unit        string `json:"unit"`
	ActiveState string `json:"active_state"`
	SubState    string `json:"sub_state"`
	Timestamp   string `json:"timestamp"`
	Hostname    string `json:"hostname"`
}

// Publisher wraps the Paho client for systemqtt topics.
type Publisher struct {
	cfg         *config.Config
	client      mqtt.Client
	onReconnect func()
}

// NewPublisher builds client options from config (does not connect).
func NewPublisher(cfg *config.Config) (*Publisher, error) {
	p := &Publisher{cfg: cfg}
	opts := mqtt.NewClientOptions()
	addrs, err := brokerAddrs(cfg)
	if err != nil {
		return nil, err
	}
	for _, a := range addrs {
		opts.AddBroker(a)
	}
	opts.SetClientID(cfg.MQTT.Client.ID)
	opts.SetUsername(cfg.MQTT.Broker.Username)
	opts.SetPassword(cfg.MQTT.Broker.Password)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)
	opts.SetConnectRetryInterval(5 * time.Second)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetCleanSession(false)

	qos := byte(cfg.MQTT.QoS)
	retained := cfg.MQTT.Retain
	statusTopic := p.statusTopic()

	opts.SetWill(statusTopic, "offline", qos, retained)

	opts.SetOnConnectHandler(func(c mqtt.Client) {
		tok := c.Publish(statusTopic, qos, retained, "online")
		_ = tok.WaitTimeout(10 * time.Second)
		if p.onReconnect != nil {
			p.onReconnect()
		}
	})

	if cfg.MQTT.TLS.Enabled {
		tc, err := newTLSConfig(cfg)
		if err != nil {
			return nil, err
		}
		opts.SetTLSConfig(tc)
	}

	p.client = mqtt.NewClient(opts)
	return p, nil
}

// SetOnReconnect sets a callback invoked after each successful connect (including reconnect).
func (p *Publisher) SetOnReconnect(fn func()) { p.onReconnect = fn }

// Connect blocks until the first connection attempt completes.
func (p *Publisher) Connect(ctx context.Context) error {
	token := p.client.Connect()
	if !token.WaitTimeout(30 * time.Second) {
		return context.DeadlineExceeded
	}
	if err := token.Error(); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func (p *Publisher) statusTopic() string {
	return fmt.Sprintf("%s/%s/status", p.cfg.MQTT.Topic.Prefix, p.cfg.Hostname)
}

func (p *Publisher) unitStateTopic(unit string) string {
	return fmt.Sprintf("%s/%s/unit/%s/state", p.cfg.MQTT.Topic.Prefix, p.cfg.Hostname, unit)
}

// PublishUnitState publishes retained JSON for a unit.
func (p *Publisher) PublishUnitState(st UnitState) error {
	b, err := json.Marshal(st)
	if err != nil {
		return err
	}
	topic := p.unitStateTopic(st.Unit)
	tok := p.client.Publish(topic, byte(p.cfg.MQTT.QoS), p.cfg.MQTT.Retain, b)
	if !tok.WaitTimeout(10 * time.Second) {
		return fmt.Errorf("publish %s: timeout", topic)
	}
	return tok.Error()
}

// PublishOffline publishes the graceful offline string (non-LWT path).
func (p *Publisher) PublishOffline() error {
	tok := p.client.Publish(p.statusTopic(), byte(p.cfg.MQTT.QoS), p.cfg.MQTT.Retain, "offline")
	if !tok.WaitTimeout(10 * time.Second) {
		return fmt.Errorf("publish offline: timeout")
	}
	return tok.Error()
}

// Disconnect closes the MQTT connection.
func (p *Publisher) Disconnect() {
	if p.client != nil && p.client.IsConnected() {
		p.client.Disconnect(250)
	}
}

// Client exposes the underlying Paho client for doctor checks.
func (p *Publisher) Client() mqtt.Client { return p.client }

func brokerAddrs(cfg *config.Config) ([]string, error) {
	raw := strings.TrimSpace(cfg.MQTT.Broker.URL)
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" {
		return nil, fmt.Errorf("mqtt.broker.url: invalid or missing host")
	}
	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = strconv.Itoa(cfg.MQTT.Broker.Port)
	}
	scheme := u.Scheme
	if cfg.MQTT.TLS.Enabled {
		scheme = "ssl"
	}
	if scheme == "" {
		scheme = "tcp"
	}
	return []string{fmt.Sprintf("%s://%s:%s", scheme, host, port)}, nil
}

func newTLSConfig(cfg *config.Config) (*tls.Config, error) {
	tc := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}
	if cfg.MQTT.TLS.Insecure {
		tc.InsecureSkipVerify = true
	}
	if ca := strings.TrimSpace(cfg.MQTT.TLS.CA); ca != "" {
		pemData, err := os.ReadFile(ca)
		if err != nil {
			return nil, fmt.Errorf("read mqtt.tls.ca: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pemData) {
			return nil, fmt.Errorf("mqtt.tls.ca: no certificates")
		}
		tc.RootCAs = pool
	}
	certFile := strings.TrimSpace(cfg.MQTT.TLS.Cert)
	keyFile := strings.TrimSpace(cfg.MQTT.TLS.Key)
	if certFile != "" && keyFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, fmt.Errorf("mqtt tls client cert: %w", err)
		}
		tc.Certificates = []tls.Certificate{cert}
	}
	return tc, nil
}
