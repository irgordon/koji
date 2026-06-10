import { describe, expect, it, vi } from 'vitest';

import { ApiError, errorMessage, fetchMetrics } from '../api';

describe('API error normalization', () => {
  it.each([
    [401, 'Authentication required', 'Sign in before using this view.'],
    [401, 'Session expired', 'Your session expired. Sign in again before continuing.'],
    [403, 'Capability denied', 'Your account does not have permission for this action.'],
    [403, 'CSRF token required', 'The request security token expired. Refresh and try again.'],
    [502, 'agent is unavailable: dial unix /run/koji/agent.sock', 'The local agent is not reachable. Approved jobs cannot run until koji-agent is running.'],
    [500, 'agent returned not implemented', 'The agent does not implement this action yet. The job cannot run on this build.'],
    [500, 'agent mutation disabled by config', 'Service mutation is disabled in agent configuration. Jobs can be approved, but they will not run until mutation is enabled intentionally.'],
    [403, 'Service is not allowlisted', 'That service is not in the configured Koji allowlist. Ask an administrator to review the service policy.'],
    [400, 'invalid service name', 'The request was rejected because one or more fields were invalid.']
  ])('maps %s %s to safe text', async (status, backendError, expected) => {
    mockJSONResponse(status, { error: backendError });

    await expect(fetchMetrics()).rejects.toMatchObject({ message: expected });
  });

  it('maps network failures to plain language', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('ECONNREFUSED 127.0.0.1')));

    await expect(fetchMetrics()).rejects.toMatchObject({
      message: 'Koji is unreachable from this browser. Check the server connection and try again.'
    });
  });

  it('maps unexpected successful payloads safely', async () => {
    mockJSONResponse(200, { sql: 'SELECT * FROM sessions' });

    await expect(fetchMetrics()).rejects.toMatchObject({
      message: 'Koji returned an unexpected response.'
    });
  });

  it('does not expose raw backend internals through errorMessage', () => {
    const error = new ApiError(500, 'unexpected_response', 'Koji returned an unexpected response.');

    expect(errorMessage(error, 'fallback')).toBe('Koji returned an unexpected response.');
    expect(errorMessage(error, 'fallback')).not.toContain('systemctl');
    expect(errorMessage(error, 'fallback')).not.toContain('SQL');
    expect(errorMessage(error, 'fallback')).not.toContain('panic');
  });
});

function mockJSONResponse(status: number, body: unknown) {
  vi.stubGlobal(
    'fetch',
    vi.fn().mockResolvedValue(
      new Response(JSON.stringify(body), {
        status,
        headers: { 'Content-Type': 'application/json' }
      })
    )
  );
}
