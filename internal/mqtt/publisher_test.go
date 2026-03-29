package mqtt

import (
	"testing"

	"github.com/systemqtt/systemqtt/internal/config"
)

func TestNewPublisher_brokerURL(t *testing.T) {
	cfg := &config.Config{}
	cfg.MQTT.Broker.URL = "tcp://mqtt.example.com"
	cfg.MQTT.Broker.Port = 1883
	cfg.MQTT.Broker.Username = "u"
	cfg.MQTT.Broker.Password = "p"
	cfg.MQTT.Client.ID = "test"
	cfg.MQTT.Topic.Prefix = "systemqtt"
	cfg.Hostname = "h1"
	cfg.Log.Level = "info"
	_, err := NewPublisher(cfg)
	if err != nil {
		t.Fatal(err)
	}
}
