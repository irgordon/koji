package system

import (
	"context"
	"errors"
	"testing"
	"time"

	"koji/internal/platform/command"
)

func TestGetServiceStatusMapsCommandFailure(t *testing.T) {
	runner := command.NewTestRunner(time.Second, 1024, []string{command.ServiceManager})

	_, err := GetServiceStatusWithRunner(context.Background(), "ssh.service", runner)
	if !errors.Is(err, command.ErrCommandFailed) {
		t.Fatalf("expected command failure, got %v", err)
	}
}

func TestParseServiceStatus(t *testing.T) {
	status := parseServiceStatus("ssh.service", "ActiveState=active\nSubState=running\n")

	if status.Name != "ssh.service" {
		t.Fatalf("unexpected service name: %s", status.Name)
	}
	if !status.Active {
		t.Fatal("expected active service")
	}
	if status.SubState != "running" {
		t.Fatalf("expected running substate, got %q", status.SubState)
	}
}

func TestGetServiceStatusPreservesNameValidation(t *testing.T) {
	runner := command.NewTestRunner(time.Second, 1024, []string{command.ServiceManager})

	_, err := GetServiceStatusWithRunner(context.Background(), "../bad", runner)
	if err == nil {
		t.Fatal("expected invalid service name")
	}
}
