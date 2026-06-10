import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';

import { Tooltip } from '../App';

describe('Tooltip', () => {
  it('renders accessible tooltip content', async () => {
    render(<Tooltip text="Koji cannot reach the local agent right now." />);

    const trigger = screen.getByLabelText('Help');
    const tooltip = screen.getByRole('tooltip');

    expect(trigger).toHaveAttribute('aria-describedby', tooltip.id);
    expect(tooltip).toHaveTextContent('Koji cannot reach the local agent right now.');

    await userEvent.tab();

    expect(trigger).toHaveFocus();
  });
});
