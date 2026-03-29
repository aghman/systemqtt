package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func getSystemdUnitStatus(unit string) (string, error) {
	cmd := exec.Command("systemctl", "is-active", unit)
	out, err := cmd.Output()
	status := strings.TrimSpace(string(out))
	if status == "" && err != nil {
		return "", fmt.Errorf("failed to get status for %s: %w", unit, err)
	}
	return status, nil
}

func main() {
	broker := os.Getenv("MQTT_BROKER")
	if broker == "" {
		broker = "tcp://localhost:1883"
	}

	units := os.Getenv("SYSTEMD_UNITS")
	if units == "" {
		units = "mosquitto"
	}

	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID("systemqtt").
		SetAutoReconnect(true)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Failed to connect to MQTT broker at %s: %v", broker, token.Error())
	}
	defer client.Disconnect(250)

	log.Printf("Connected to MQTT broker at %s", broker)

	unitList := strings.Split(units, ",")
	for _, unit := range unitList {
		unit = strings.TrimSpace(unit)
		if unit == "" {
			continue
		}

		status, err := getSystemdUnitStatus(unit)
		if err != nil {
			log.Printf("Warning: %v", err)
			status = "unknown"
		}

		topic := fmt.Sprintf("systemqtt/%s/status", unit)
		token := client.Publish(topic, 0, false, status)
		token.Wait()

		if token.Error() != nil {
			log.Printf("Failed to publish status for %s: %v", unit, token.Error())
		} else {
			log.Printf("Published %s = %s", topic, status)
		}
	}

	time.Sleep(500 * time.Millisecond)
	log.Println("Done.")
}
