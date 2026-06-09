package system

import (
	"context"
	"errors"
	"strings"

	"koji/internal/platform/command"
)

type ServiceStatus struct {
	Name     string `json:"name"`
	Active   bool   `json:"active"`
	SubState string `json:"subState"`
}

func GetServiceStatus(ctx context.Context, name string) (ServiceStatus, error) {
	return GetServiceStatusWithRunner(ctx, name, command.NewReadOnlyRunner())
}

func GetServiceStatusWithRunner(ctx context.Context, name string, runner command.Runner) (ServiceStatus, error) {
	if err := ValidateServiceName(name); err != nil {
		return ServiceStatus{}, err
	}

	result, err := runner.Run(ctx, "systemctl", "show", name, "--property=ActiveState,SubState")
	if err != nil {
		return ServiceStatus{}, command.SafeError(err)
	}

	return parseServiceStatus(name, string(result.Stdout)), nil
}

func parseServiceStatus(name string, output string) ServiceStatus {
	var activeState, subState string
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
			switch parts[0] {
			case "ActiveState":
				activeState = parts[1]
			case "SubState":
				subState = parts[1]
			}
		}
	}

	return ServiceStatus{
		Name:     name,
		Active:   activeState == "active",
		SubState: subState,
	}
}

func ValidateServiceName(name string) error {
	if len(name) == 0 || len(name) > 64 {
		return errors.New("invalid unit name length")
	}
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '.' || r == '_') {
			return errors.New("invalid character in unit name")
		}
	}
	return nil
}
