package jobs

import (
	"context"
	"errors"
	"strings"
	"time"

	"koji/internal/agent"
	"koji/internal/audit"
)

const (
	DefaultWorkerPollInterval = time.Second
	ReasonAgentUnavailable    = "agent_unavailable"
	ReasonCommandFailed       = "command_failed"
	ReasonCommandTimeout      = "command_timeout"
	ReasonInvalidJob          = "invalid_job"
	ReasonMutationDisabled    = "mutation_disabled"
	ReasonServiceDenied       = "service_not_allowlisted"
	ReasonUnsupportedAction   = "unsupported_action"
	ReasonValidationError     = "validation_error"
)

type Worker struct {
	store        *Store
	agent        agent.ServiceController
	audit        *audit.Store
	pollInterval time.Duration
}

func NewWorker(store *Store, agent agent.ServiceController, auditStore *audit.Store, pollInterval time.Duration) *Worker {
	return &Worker{
		store:        store,
		agent:        agent,
		audit:        auditStore,
		pollInterval: effectivePollInterval(pollInterval),
	}
}

func (w *Worker) Start(ctx context.Context) error {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		if ctx.Err() != nil {
			return nil
		}
		if err := w.ProcessOne(ctx); err != nil && !errors.Is(err, ErrNoApprovedJobs) {
			return err
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (w *Worker) ProcessOne(ctx context.Context) error {
	job, err := w.store.ClaimApproved(ctx)
	if err != nil {
		return err
	}
	if err := w.recordJobEvent(ctx, audit.ActionJobStarted, audit.OutcomeAccepted, job); err != nil {
		return err
	}
	return w.executeClaimedJob(ctx, job)
}

func (w *Worker) executeClaimedJob(ctx context.Context, job Job) error {
	request, err := agentRequestForJob(job)
	if err != nil {
		return w.markFailed(ctx, job.ID, ReasonInvalidJob)
	}
	if err := w.agent.ControlService(ctx, request); err != nil {
		return w.applyAgentError(ctx, job.ID, err)
	}
	return w.markFailed(ctx, job.ID, ReasonInvalidJob)
}

func (w *Worker) applyAgentError(ctx context.Context, id string, err error) error {
	if errors.Is(err, agent.ErrNotImplemented) {
		return w.markNotImplemented(ctx, id, StatusNotImplemented)
	}
	if errors.Is(err, agent.ErrMutationDisabled) {
		return w.markNotImplemented(ctx, id, ReasonMutationDisabled)
	}
	if errors.Is(err, agent.ErrCommandFailed) {
		return w.markFailed(ctx, id, ReasonCommandFailed)
	}
	if errors.Is(err, agent.ErrAgentTimeout) {
		return w.markFailed(ctx, id, ReasonCommandTimeout)
	}
	if errors.Is(err, agent.ErrServiceNotAllowlisted) {
		return w.markFailed(ctx, id, ReasonServiceDenied)
	}
	if errors.Is(err, agent.ErrUnsupportedServiceAction) {
		return w.markFailed(ctx, id, ReasonUnsupportedAction)
	}
	if errors.Is(err, agent.ErrInvalidServiceName) {
		return w.markFailed(ctx, id, ReasonValidationError)
	}
	return w.markFailed(ctx, id, ReasonAgentUnavailable)
}

func (w *Worker) markNotImplemented(ctx context.Context, id string, reason string) error {
	job, err := w.store.markNotImplemented(ctx, id, reason)
	if err != nil {
		return err
	}
	if err := w.recordJobEvent(ctx, audit.ActionJobNotImpl, audit.OutcomeRejected, job); err != nil {
		return err
	}
	return w.recordJobEvent(ctx, audit.ActionJobStatus, audit.OutcomeAccepted, job)
}

func (w *Worker) markFailed(ctx context.Context, id string, reason string) error {
	job, err := w.store.MarkFailed(ctx, id, reason)
	if err != nil {
		return err
	}
	if err := w.recordJobEvent(ctx, audit.ActionJobFailed, audit.OutcomeFailure, job); err != nil {
		return err
	}
	return w.recordJobEvent(ctx, audit.ActionJobStatus, audit.OutcomeAccepted, job)
}

func (w *Worker) recordJobEvent(ctx context.Context, action string, outcome string, job Job) error {
	return w.audit.Record(ctx, audit.Event{
		Action:     action,
		Target:     "jobs:" + job.ID,
		Outcome:    outcome,
		ReasonCode: job.StatusReason,
		RequestID:  job.RequestID,
	})
}

func agentRequestForJob(job Job) (agent.ServiceControlRequest, error) {
	action, ok := strings.CutPrefix(job.Action, "service.")
	if !ok {
		return agent.ServiceControlRequest{}, ReasonError(ReasonInvalidJob)
	}
	return agent.ServiceControlRequest{
		Service: job.Target,
		Action:  action,
	}, nil
}

func effectivePollInterval(interval time.Duration) time.Duration {
	if interval <= 0 {
		return DefaultWorkerPollInterval
	}
	return interval
}

type ReasonError string

func (e ReasonError) Error() string {
	return string(e)
}
