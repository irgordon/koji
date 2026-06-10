import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { ControlPlaneObservability } from '../App';
import type { HealthResponse, ObservabilityMetrics } from '../types';

describe('ControlPlaneObservability', () => {
  it('renders job, worker, agent, audit, and auth metrics', () => {
    render(<ControlPlaneObservability metrics={metrics()} readiness={readiness('ok')} />);

    expect(screen.getByText('Job Flow')).toBeInTheDocument();
    expect(screen.getByText('7 created')).toBeInTheDocument();
    expect(screen.getByText('Worker')).toBeInTheDocument();
    expect(screen.getByText('Agent RPC')).toBeInTheDocument();
    expect(screen.getByText('Audit Writes')).toBeInTheDocument();
    expect(screen.getByText('Authentication')).toBeInTheDocument();
    expect(screen.getByText('OK Agent reachable')).toBeInTheDocument();
  });

  it('fails gracefully when metrics are missing', () => {
    render(<ControlPlaneObservability metrics={null} readiness={null} />);

    expect(screen.getByRole('status')).toHaveTextContent('Loading control-plane metrics');
  });

  it('marks agent degraded when readiness reports degraded', () => {
    render(<ControlPlaneObservability metrics={metrics()} readiness={readiness('degraded')} />);

    expect(screen.getByText('WARN Agent unavailable')).toBeInTheDocument();
  });
});

function metrics(): ObservabilityMetrics {
  return {
    counters: {
      jobs_created_total: 7,
      worker_polls_total: 12,
      worker_errors_total: 0,
      agent_rpc_requests_total: 3,
      agent_rpc_failures_total: 0,
      audit_writes_total: 9,
      audit_write_failures_total: 0,
      auth_login_success_total: 4,
      auth_login_failure_total: 1,
      readiness_checks_total: 5,
      readiness_db_failures_total: 0,
      readiness_migration_failures_total: 0,
      readiness_agent_degraded_total: 0
    },
    jobs_by_status: {
      queued: 2,
      running: 1,
      failed: 0
    }
  };
}

function readiness(agentStatus: 'ok' | 'degraded'): HealthResponse {
  return {
    status: agentStatus,
    checks: {
      agent: { status: agentStatus }
    }
  };
}
