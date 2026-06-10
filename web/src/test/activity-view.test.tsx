import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { ActivityView } from '../App';
import type { ActivityEvent } from '../types';

describe('ActivityView', () => {
  it('renders empty state', () => {
    render(<ActivityView events={[]} error={null} updatedAt={null} />);

    expect(screen.getByText('No activity available')).toBeInTheDocument();
  });

  it('renders activity rows with request IDs', () => {
    render(<ActivityView events={[activityEvent()]} error={null} updatedAt={new Date()} />);

    expect(screen.getByText('Login')).toBeInTheDocument();
    expect(screen.getByText('OK Success')).toBeInTheDocument();
    expect(screen.getByText('request-abc')).toBeInTheDocument();
  });

  it('renders errors safely', () => {
    render(<ActivityView events={[]} error="Your account does not have permission for this action." updatedAt={null} />);

    expect(screen.getByRole('alert')).toHaveTextContent('Your account does not have permission for this action.');
  });
});

function activityEvent(): ActivityEvent {
  return {
    timestamp: '2026-06-10T12:00:00Z',
    action: 'auth.login',
    target: 'session',
    outcome: 'success',
    reason_code: 'session_created',
    request_id: 'request-abc'
  };
}
