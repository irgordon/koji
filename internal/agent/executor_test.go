package agent

import (
	"context"
	"errors"
	"testing"
	"time"

	"koji/internal/config"
	"koji/internal/platform/command"
)

type recordingRunner struct {
	err        error
	executable string
	args       []string
	called     bool
}

func (r *recordingRunner) Run(ctx context.Context, executable string, args ...string) (command.Result, error) {
	r.called = true
	r.executable = executable
	r.args = append([]string(nil), args...)
	return command.Result{}, r.err
}

func TestExecutorReturnsMutationDisabledByDefault(t *testing.T) {
	executor := NewExecutor(config.NewDefaultConfig())

	response := executor.ControlService(context.Background(), ServiceControlRequest{
		Service: "ssh.service",
		Action:  "restart",
	})

	assertAgentResponse(t, response, ResponseError, ReasonMutationDisabled)
}

func TestExecutorRejectsNonAllowlistedService(t *testing.T) {
	executor := enabledExecutor([]string{"kojid.service"}, nil)

	response := executor.ControlService(context.Background(), ServiceControlRequest{
		Service: "ssh.service",
		Action:  "restart",
	})

	assertAgentResponse(t, response, ResponseError, ReasonServiceNotAllowlisted)
}

func TestExecutorRejectsUnsupportedAction(t *testing.T) {
	executor := enabledExecutor([]string{"ssh.service"}, nil)

	response := executor.ControlService(context.Background(), ServiceControlRequest{
		Service: "ssh.service",
		Action:  "reload",
	})

	assertAgentResponse(t, response, ResponseError, ReasonUnsupportedAction)
}

func TestEnabledMutationUsesPlatformCommandRunner(t *testing.T) {
	runner := &recordingRunner{}
	executor := enabledExecutor([]string{"ssh.service"}, runner)

	response := executor.ControlService(context.Background(), ServiceControlRequest{
		Service: "ssh.service",
		Action:  "restart",
	})

	assertAgentResponse(t, response, ResponseOK, "")
	if !runner.called {
		t.Fatal("expected command runner call")
	}
	if runner.executable != "systemctl" {
		t.Fatalf("expected systemctl executable, got %q", runner.executable)
	}
	if len(runner.args) != 2 || runner.args[0] != "restart" || runner.args[1] != "ssh.service" {
		t.Fatalf("unexpected runner args: %#v", runner.args)
	}
}

func TestExecutorMapsCommandTimeout(t *testing.T) {
	executor := enabledExecutor([]string{"ssh.service"}, &recordingRunner{err: command.ErrCommandTimeout})

	response := executor.ControlService(context.Background(), ServiceControlRequest{
		Service: "ssh.service",
		Action:  "restart",
	})

	assertAgentResponse(t, response, ResponseError, ReasonCommandTimeout)
}

func TestClientMapsMutationDisabled(t *testing.T) {
	err := responseError(RPCResponse{Status: ResponseError, Reason: ReasonMutationDisabled})

	if !errors.Is(err, ErrMutationDisabled) {
		t.Fatalf("expected mutation disabled, got %v", err)
	}
}

func enabledExecutor(allowlist []string, runner commandRunner) Executor {
	if runner == nil {
		runner = &recordingRunner{}
	}
	return NewExecutorWithRunner(config.Config{
		AgentMutationEnabled:    true,
		AgentServiceAllowlist:   allowlist,
		AgentCommandTimeout:     time.Second,
		AgentCommandOutputLimit: command.DefaultOutputLimit,
	}, runner)
}

func assertAgentResponse(t *testing.T, response RPCResponse, status string, reason string) {
	t.Helper()

	if response.Status != status || response.Reason != reason {
		t.Fatalf("expected %s/%s, got %#v", status, reason, response)
	}
}
