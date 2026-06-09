package agent

import (
	"context"
	"errors"
	"time"

	"koji/internal/config"
	"koji/internal/platform/command"
	"koji/internal/system"
)

type Executor struct {
	mutationEnabled bool
	allowlist       map[string]struct{}
	runner          command.CommandRunner
}

func NewExecutor(cfg config.Config) Executor {
	return NewExecutorWithRunner(cfg, command.NewAgentMutationRunner(cfg.AgentCommandTimeout, cfg.AgentCommandOutputLimit))
}

func NewExecutorWithRunner(cfg config.Config, runner command.CommandRunner) Executor {
	return Executor{
		mutationEnabled: cfg.AgentMutationEnabled,
		allowlist:       serviceAllowlist(cfg.AgentServiceAllowlist),
		runner:          runner,
	}
}

func (e Executor) ControlService(ctx context.Context, request ServiceControlRequest) RPCResponse {
	if !e.mutationEnabled {
		return errorResponse(ReasonMutationDisabled)
	}
	if reason := validateExecutorRequest(request); reason != "" {
		return errorResponse(reason)
	}
	if !e.serviceAllowed(request.Service) {
		return errorResponse(ReasonServiceNotAllowlisted)
	}
	if err := e.runServiceMutation(ctx, request); err != nil {
		return errorResponse(commandFailureReason(err))
	}
	return RPCResponse{Status: ResponseOK}
}

func (e Executor) serviceAllowed(service string) bool {
	_, ok := e.allowlist[service]
	return ok
}

func (e Executor) runServiceMutation(ctx context.Context, request ServiceControlRequest) error {
	_, err := command.RunServiceMutation(ctx, e.runner, request.Action, request.Service)
	return err
}

func validateExecutorRequest(request ServiceControlRequest) string {
	if err := system.ValidateServiceName(request.Service); err != nil {
		return ReasonValidationError
	}
	if !isSupportedServiceAction(request.Action) {
		return ReasonUnsupportedAction
	}
	return ""
}

func commandFailureReason(err error) string {
	switch {
	case errors.Is(err, command.ErrCommandTimeout):
		return ReasonCommandTimeout
	default:
		return ReasonCommandFailed
	}
}

func errorResponse(reason string) RPCResponse {
	return RPCResponse{Status: ResponseError, Reason: reason}
}

func serviceAllowlist(services []string) map[string]struct{} {
	allowlist := make(map[string]struct{}, len(services))
	for _, service := range services {
		allowlist[service] = struct{}{}
	}
	return allowlist
}

func DisabledExecutor() Executor {
	return NewExecutorWithRunner(config.Config{
		AgentMutationEnabled:    false,
		AgentCommandTimeout:     time.Second,
		AgentCommandOutputLimit: command.DefaultOutputLimit,
		AgentServiceAllowlist:   nil,
	}, command.NewAgentMutationRunner(time.Second, command.DefaultOutputLimit))
}
