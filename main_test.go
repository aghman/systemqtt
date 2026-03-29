package main

import (
	"testing"
)

func TestGetSystemdUnitStatus(t *testing.T) {
	status, err := getSystemdUnitStatus("mosquitto")
	if err != nil {
		t.Logf("Warning: could not get mosquitto status (may not be running): %v", err)
	}
	t.Logf("mosquitto status: %s", status)

	validStatuses := map[string]bool{
		"active":       true,
		"inactive":     true,
		"failed":       true,
		"activating":   true,
		"deactivating": true,
		"unknown":      true,
	}

	if status != "" && !validStatuses[status] {
		t.Errorf("unexpected status value: %q", status)
	}
}

func TestGetSystemdUnitStatusInvalidUnit(t *testing.T) {
	status, _ := getSystemdUnitStatus("nonexistent-unit-12345")
	validStatuses := map[string]bool{
		"inactive": true,
		"unknown":  true,
		"":         true,
	}
	if !validStatuses[status] {
		t.Errorf("expected inactive/unknown for nonexistent unit, got: %q", status)
	}
}
