import { fireEvent, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { act } from 'react';
import { describe, expect, it, vi } from 'vitest';

import { ToastProvider, useToast } from '../App';
import type { ToastRequest } from '../App';

function ToastHarness({ toast }: { toast: ToastRequest }) {
  const { notify } = useToast();
  return <button onClick={() => notify(toast)}>Notify</button>;
}

function renderToast(toast: ToastRequest) {
  return render(
    <ToastProvider>
      <ToastHarness toast={toast} />
    </ToastProvider>
  );
}

describe('ToastProvider', () => {
  it.each([
    ['success', 'Saved'],
    ['error', 'Failed'],
    ['warning', 'Needs review'],
    ['info', 'Heads up']
  ] as const)('renders %s toast feedback', async (type, title) => {
    renderToast({ type, title, message: 'Operator-facing message.' });

    await userEvent.click(screen.getByRole('button', { name: 'Notify' }));

    expect(screen.getByText(title)).toBeInTheDocument();
    expect(screen.getByText('Operator-facing message.')).toBeInTheDocument();
    expect(screen.getByLabelText('Notifications')).toHaveAttribute('aria-live', 'polite');
  });

  it('dismisses a toast manually', async () => {
    renderToast({ type: 'error', title: 'Failed', message: 'Action could not complete.' });

    await userEvent.click(screen.getByRole('button', { name: 'Notify' }));
    await userEvent.click(screen.getByRole('button', { name: 'Dismiss Failed' }));

    expect(screen.queryByText('Failed')).not.toBeInTheDocument();
  });

  it('auto-dismisses success toasts only', async () => {
    vi.useFakeTimers();
    render(
      <ToastProvider>
        <ToastHarness toast={{ type: 'success', title: 'Done', message: 'Action completed.' }} />
        <ToastHarness toast={{ type: 'error', title: 'Still here', message: 'Manual dismissal required.' }} />
      </ToastProvider>
    );

    const buttons = screen.getAllByRole('button', { name: 'Notify' });
    fireEvent.click(buttons[0]);
    fireEvent.click(buttons[1]);

    act(() => vi.advanceTimersByTime(5000));

    expect(screen.queryByText('Done')).not.toBeInTheDocument();
    expect(screen.getByText('Still here')).toBeInTheDocument();
  });
});
