package at

import (
	"testing"
	"time"
)

func TestGetSIMStatus(t *testing.T) {
	for _, status := range []string{"READY", "SIM PIN", "SIM PUK", "PH-SIM PIN"} {
		t.Run(status, func(t *testing.T) {
			device := newCommandTestDevice(time.Second, func(command string, responses chan<- string) error {
				if command != "AT+CPIN?\r\n" {
					t.Errorf("unexpected SIM query: %q", command)
				}
				responses <- "AT+CPIN?"
				responses <- "+CPIN: " + status
				responses <- "OK"
				return nil
			})
			got, err := device.GetSIMStatus()
			if err != nil || got != status {
				t.Fatalf("expected status %q, got %q, err=%v", status, got, err)
			}
		})
	}
}

func TestGetSIMStatusRejectsMissingStatus(t *testing.T) {
	device := newCommandTestDevice(time.Second, func(command string, responses chan<- string) error {
		responses <- "ERROR"
		return nil
	})
	if status, err := device.GetSIMStatus(); err == nil || status != "" {
		t.Fatalf("missing SIM status was accepted: status=%q, err=%v", status, err)
	}
}
