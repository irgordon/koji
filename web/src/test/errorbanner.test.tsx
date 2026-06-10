import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { ErrorBanner } from '../App';

describe('ErrorBanner', () => {
  it('renders plain-language alert text', () => {
    render(<ErrorBanner message="Your account does not have permission for this action." />);

    const alert = screen.getByRole('alert');
    expect(alert).toHaveTextContent('Your account does not have permission for this action.');
  });

  it('renders unsafe backend text as text, not HTML', () => {
    render(<ErrorBanner message="<img src=x onerror=alert(1)> SQL error" />);

    expect(screen.getByText('<img src=x onerror=alert(1)> SQL error')).toBeInTheDocument();
    expect(document.querySelector('img')).not.toBeInTheDocument();
  });
});
