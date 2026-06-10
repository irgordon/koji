import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { JobsView } from '../App';
import type { JobRecord, JobStatus } from '../types';

describe('JobsView', () => {
  it.each([
    ['queued', 'WARN Waiting for approval'],
    ['approved', 'OK Approved, waiting for worker'],
    ['running', 'WARN Running'],
    ['completed', 'OK Completed'],
    ['failed', 'FAIL Failed'],
    ['rejected', 'FAIL Rejected'],
    ['not_implemented', 'WARN Agent not implemented']
  ] as Array<[JobStatus, string]>)('renders human-readable status for %s', (status, label) => {
    render(<JobsView jobs={[job(status)]} error={null} updatedAt={new Date()} onDecision={vi.fn()} />);

    expect(screen.getByText(label)).toBeInTheDocument();
  });

  it('shows approval controls only for queued jobs', () => {
    render(
      <JobsView
        jobs={[job('queued'), job('approved', 'approved-job')]}
        error={null}
        updatedAt={new Date()}
        onDecision={vi.fn()}
      />
    );

    expect(screen.getByRole('button', { name: 'Approve' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Reject' })).toBeInTheDocument();
    expect(screen.getByLabelText('Decision reason for queued-job')).toBeInTheDocument();
    expect(screen.queryByLabelText('Decision reason for approved-job')).not.toBeInTheDocument();
  });

  it('renders job page errors as alerts', () => {
    render(<JobsView jobs={[]} error="Your account does not have permission for this action." updatedAt={null} onDecision={vi.fn()} />);

    expect(screen.getByRole('alert')).toHaveTextContent('Your account does not have permission for this action.');
  });
});

function job(status: JobStatus, id = `${status}-job`): JobRecord {
  return {
    id,
    created_at: '2026-06-10T12:00:00Z',
    created_by: 1,
    action: 'service.restart',
    target: 'ssh.service',
    status,
    status_reason: status,
    request_id: 'request-123',
    decision_reason: ''
  };
}
