import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import App from '../App';

describe('responsive shell smoke checks', () => {
  it('keeps navigation and critical controls mounted at mobile width', async () => {
    setViewportWidth(375);
    mockAppFetch();

    render(<App />);

    expect(await screen.findByRole('button', { name: 'Overview' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Services' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Processes' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Jobs' })).toBeInTheDocument();
    expect(screen.getByLabelText('Primary navigation')).toBeInTheDocument();
  });
});

function setViewportWidth(width: number) {
  Object.defineProperty(window, 'innerWidth', { configurable: true, value: width });
  Object.defineProperty(document.documentElement, 'clientWidth', { configurable: true, value: width });
  window.dispatchEvent(new Event('resize'));
}

function mockAppFetch() {
  vi.stubGlobal('fetch', vi.fn((input: RequestInfo | URL) => {
    const url = String(input);
    return Promise.resolve(
      new Response(JSON.stringify(payloadForURL(url)), {
        status: 200,
        headers: { 'Content-Type': 'application/json' }
      })
    );
  }));
}

function payloadForURL(url: string): unknown {
  if (url.includes('/api/v1/metrics')) {
    return { cpuUsage: 1, memTotal: 100, memAvailable: 90, memUsed: 10, memUsagePct: 10, uptime: 1000 };
  }
  if (url.includes('/api/v1/disk')) {
    return { path: '/', totalBytes: 1000, freeBytes: 800, usedBytes: 200, usagePct: 20 };
  }
  if (url.includes('/healthz') || url.includes('/readyz')) {
    return { status: 'ok', checks: { agent: { status: 'ok' } } };
  }
  if (url.includes('/api/observability/metrics')) {
    return { counters: {}, jobs_by_status: {} };
  }
  return {};
}
