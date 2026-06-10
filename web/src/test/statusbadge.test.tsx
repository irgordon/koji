import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { StatusBadge } from '../App';
import type { UIStatus } from '../App';

describe('StatusBadge', () => {
  it.each([
    ['ok', 'OK Healthy'],
    ['degraded', 'WARN Degraded'],
    ['fail', 'FAIL Failed'],
    ['running', 'RUN Running'],
    ['completed', 'DONE Completed'],
    ['pending', 'WAIT Pending']
  ] as Array<[UIStatus, string]>)('renders non-color-only label for %s', (status, label) => {
    render(<StatusBadge status={status} />);

    expect(screen.getByText(label)).toBeInTheDocument();
  });

  it('uses explicit labels when provided', () => {
    render(<StatusBadge status="degraded" label="Needs approval" />);

    expect(screen.getByText('WARN Needs approval')).toBeInTheDocument();
  });
});
