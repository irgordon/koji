import type {
  ActivityEvent,
  ActivityResponse,
  DiskMetrics,
  HealthResponse,
  HealthStatus,
  JobRecord,
  JobsResponse,
  JobStatus,
  NormalizedErrorCode,
  ObservabilityMetrics,
  ProcessInfo,
  ServiceControlAction,
  ServiceControlJobResponse,
  ServiceListResponse,
  ServiceStatus,
  SystemMetrics
} from './types';

export class ApiError extends Error {
  readonly status: number;
  readonly code: NormalizedErrorCode;

  constructor(status: number, code: NormalizedErrorCode, message: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
  }
}

export async function fetchMetrics(signal?: AbortSignal): Promise<SystemMetrics> {
  return requestJSON<SystemMetrics>('/api/v1/metrics', { signal }, isSystemMetrics);
}

export async function fetchDisk(signal?: AbortSignal): Promise<DiskMetrics> {
  return requestJSON<DiskMetrics>('/api/v1/disk', { signal }, isDiskMetrics);
}

export async function fetchServices(signal?: AbortSignal): Promise<ServiceStatus[]> {
  const response = await requestJSON<ServiceListResponse>('/api/v1/services', { signal }, isServiceListResponse);
  return response.services;
}

export async function fetchProcesses(signal?: AbortSignal): Promise<ProcessInfo[]> {
  return requestJSON<ProcessInfo[]>('/api/v1/processes', { signal }, isProcessList);
}

export async function fetchHealth(signal?: AbortSignal): Promise<HealthResponse> {
  return requestJSON<HealthResponse>('/healthz', { signal }, isHealthResponse);
}

export async function fetchReadiness(signal?: AbortSignal): Promise<HealthResponse> {
  return requestJSON<HealthResponse>('/readyz', { signal }, isHealthResponse);
}

export async function fetchActivity(signal?: AbortSignal): Promise<ActivityEvent[]> {
  const response = await requestJSON<ActivityResponse>('/api/activity', { signal }, isActivityResponse);
  return response.events;
}

export async function fetchJobs(signal?: AbortSignal): Promise<JobRecord[]> {
  const response = await requestJSON<JobsResponse>('/api/jobs', { signal }, isJobsResponse);
  return response.jobs;
}

export async function fetchObservabilityMetrics(signal?: AbortSignal): Promise<ObservabilityMetrics> {
  return requestJSON<ObservabilityMetrics>(
    '/api/observability/metrics',
    { signal },
    isObservabilityMetrics
  );
}

export async function approveJob(id: string, reason: string): Promise<JobRecord> {
  return decideJob(id, 'approve', reason);
}

export async function rejectJob(id: string, reason: string): Promise<JobRecord> {
  return decideJob(id, 'reject', reason);
}

export async function controlService(service: string, action: ServiceControlAction): Promise<ServiceControlJobResponse> {
  return requestJSON<ServiceControlJobResponse>(
    `/api/services/${encodeURIComponent(service)}/${encodeURIComponent(action)}`,
    {
      method: 'POST',
      headers: csrfHeaders()
    },
    isServiceControlJobResponse
  );
}

export function errorMessage(error: unknown, fallback: string): string {
  if (isAbortError(error)) {
    return fallback;
  }
  if (error instanceof ApiError || error instanceof Error) {
    return error.message;
  }
  return fallback;
}

export function isAbortError(error: unknown): boolean {
  return error instanceof DOMException && error.name === 'AbortError';
}

async function requestJSON<T>(url: string, init: RequestInit, guard: (value: unknown) => value is T): Promise<T> {
  const response = await fetchSafely(url, init);
  const payload: unknown = await readPayload(response);

  if (!response.ok) {
    throw normalizedHTTPError(response.status, payload);
  }
  if (!guard(payload)) {
    throw new ApiError(response.status, 'unexpected_response', errorText('unexpected_response'));
  }
  return payload;
}

async function fetchSafely(url: string, init: RequestInit): Promise<Response> {
  try {
    return await fetch(url, init);
  } catch (error: unknown) {
    if (isAbortError(error)) {
      throw error;
    }
    throw new ApiError(0, 'network_error', errorText('network_error'));
  }
}

async function readPayload(response: Response): Promise<unknown> {
  const text = await response.text();
  if (text === '') {
    return {};
  }
  try {
    return JSON.parse(text) as unknown;
  } catch (error) {
    throw new ApiError(response.status, 'unexpected_response', errorText('unexpected_response'));
  }
}

function normalizedHTTPError(status: number, payload: unknown): ApiError {
  const backendMessage = errorPayloadMessage(payload);
  const code = errorCode(status, backendMessage);
  return new ApiError(status, code, errorText(code));
}

function errorPayloadMessage(payload: unknown): string {
  return isRecord(payload) && typeof payload.error === 'string' ? payload.error : '';
}

function errorCode(status: number, message: string): NormalizedErrorCode {
  const normalized = message.toLowerCase();
  if (status === 401) {
    return normalized.includes('expired') || normalized.includes('invalid session')
      ? 'session_expired'
      : 'unauthenticated';
  }
  if (normalized.includes('csrf')) {
    return 'csrf_missing_or_invalid';
  }
  if (normalized.includes('mutation disabled')) {
    return 'mutation_disabled';
  }
  if (normalized.includes('not allowlisted')) {
    return 'service_not_allowlisted';
  }
  if (normalized.includes('not implemented')) {
    return 'agent_not_implemented';
  }
  if (normalized.includes('agent is unavailable') || status === 502) {
    return 'agent_unavailable';
  }
  if (status === 409) {
    return 'job_conflict';
  }
  if (status === 403) {
    return 'forbidden';
  }
  if (status === 400) {
    return 'validation_error';
  }
  return 'unexpected_response';
}

function errorText(code: NormalizedErrorCode): string {
  switch (code) {
    case 'unauthenticated':
      return 'Sign in before using this view.';
    case 'forbidden':
      return 'Your account does not have permission for this action.';
    case 'csrf_missing_or_invalid':
      return 'The request security token expired. Refresh and try again.';
    case 'agent_unavailable':
      return 'The local Koji agent is unavailable, so privileged service actions cannot run.';
    case 'agent_not_implemented':
      return 'Service control is not enabled in this build yet.';
    case 'mutation_disabled':
      return 'Service mutation is disabled by Koji configuration.';
    case 'service_not_allowlisted':
      return 'That service is not in the configured Koji allowlist.';
    case 'job_conflict':
      return 'That job is no longer queued, so it cannot be approved or rejected.';
    case 'validation_error':
      return 'The request was rejected because one or more fields were invalid.';
    case 'network_error':
      return 'Koji is unreachable from the browser right now.';
    case 'unexpected_response':
      return 'Koji returned an unexpected response.';
    case 'session_expired':
      return 'Your session expired. Sign in again before continuing.';
  }
}

function csrfHeaders(): HeadersInit {
  const token = cookieValue('koji_csrf');
  return token ? { 'X-CSRF-Token': token } : {};
}

function jsonHeaders(): HeadersInit {
  return {
    ...csrfHeaders(),
    'Content-Type': 'application/json'
  };
}

function decideJob(id: string, decision: 'approve' | 'reject', reason: string): Promise<JobRecord> {
  return requestJSON<JobRecord>(
    `/api/jobs/${encodeURIComponent(id)}/${decision}`,
    {
      method: 'POST',
      headers: jsonHeaders(),
      body: JSON.stringify({ reason })
    },
    isJobRecord
  );
}

function cookieValue(name: string): string | null {
  const prefix = `${name}=`;
  const entry = document.cookie.split('; ').find((cookie) => cookie.startsWith(prefix));
  return entry ? decodeURIComponent(entry.slice(prefix.length)) : null;
}

function isSystemMetrics(value: unknown): value is SystemMetrics {
  return (
    isRecord(value) &&
    isNumber(value.cpuUsage) &&
    isNumber(value.memTotal) &&
    isNumber(value.memAvailable) &&
    isNumber(value.memUsed) &&
    isNumber(value.memUsagePct) &&
    isNumber(value.uptime)
  );
}

function isDiskMetrics(value: unknown): value is DiskMetrics {
  return (
    isRecord(value) &&
    typeof value.path === 'string' &&
    isNumber(value.totalBytes) &&
    isNumber(value.freeBytes) &&
    isNumber(value.usedBytes) &&
    isNumber(value.usagePct)
  );
}

function isServiceListResponse(value: unknown): value is ServiceListResponse {
  return isRecord(value) && Array.isArray(value.services) && value.services.every(isServiceStatus);
}

function isServiceStatus(value: unknown): value is ServiceStatus {
  return (
    isRecord(value) &&
    typeof value.name === 'string' &&
    typeof value.active === 'boolean' &&
    typeof value.subState === 'string'
  );
}

function isProcessList(value: unknown): value is ProcessInfo[] {
  return Array.isArray(value) && value.every(isProcessInfo);
}

function isProcessInfo(value: unknown): value is ProcessInfo {
  return (
    isRecord(value) &&
    isNumber(value.pid) &&
    typeof value.name === 'string' &&
    typeof value.state === 'string' &&
    isOptionalNumber(value.uid) &&
    isOptionalNumber(value.ppid) &&
    isOptionalNumber(value.cpuUser) &&
    isOptionalNumber(value.cpuSystem) &&
    isOptionalNumber(value.rss) &&
    isOptionalNumber(value.memoryPct) &&
    isOptionalString(value.commandLine)
  );
}

function isHealthResponse(value: unknown): value is HealthResponse {
  return isRecord(value) && isHealthStatus(value.status) && isHealthChecks(value.checks);
}

function isHealthChecks(value: unknown): value is Record<string, { status: HealthStatus }> {
  return isRecord(value) && Object.values(value).every(isHealthCheck);
}

function isHealthCheck(value: unknown): value is { status: HealthStatus } {
  return isRecord(value) && isHealthStatus(value.status);
}

function isHealthStatus(value: unknown): value is HealthStatus {
  return value === 'ok' || value === 'degraded' || value === 'fail';
}

function isActivityResponse(value: unknown): value is ActivityResponse {
  return isRecord(value) && Array.isArray(value.events) && value.events.every(isActivityEvent);
}

function isActivityEvent(value: unknown): value is ActivityEvent {
  return (
    isRecord(value) &&
    typeof value.timestamp === 'string' &&
    typeof value.action === 'string' &&
    typeof value.target === 'string' &&
    typeof value.outcome === 'string' &&
    typeof value.reason_code === 'string' &&
    typeof value.request_id === 'string'
  );
}

function isJobsResponse(value: unknown): value is JobsResponse {
  return isRecord(value) && Array.isArray(value.jobs) && value.jobs.every(isJobRecord);
}

function isJobRecord(value: unknown): value is JobRecord {
  return (
    isRecord(value) &&
    typeof value.id === 'string' &&
    typeof value.created_at === 'string' &&
    isOptionalNumber(value.created_by) &&
    typeof value.action === 'string' &&
    typeof value.target === 'string' &&
    isJobStatus(value.status) &&
    typeof value.status_reason === 'string' &&
    typeof value.request_id === 'string' &&
    isOptionalNumber(value.approved_by) &&
    isOptionalString(value.approved_at) &&
    isOptionalNumber(value.rejected_by) &&
    isOptionalString(value.rejected_at) &&
    typeof value.decision_reason === 'string' &&
    isOptionalString(value.started_at)
  );
}

function isJobStatus(value: unknown): value is JobStatus {
  return (
    value === 'queued' ||
    value === 'approved' ||
    value === 'rejected' ||
    value === 'running' ||
    value === 'completed' ||
    value === 'failed' ||
    value === 'not_implemented'
  );
}

function isServiceControlJobResponse(value: unknown): value is ServiceControlJobResponse {
  return isRecord(value) && typeof value.jobId === 'string' && isJobStatus(value.status);
}

function isObservabilityMetrics(value: unknown): value is ObservabilityMetrics {
  return isRecord(value) && isNumberRecord(value.counters) && isNumberRecord(value.jobs_by_status);
}

function isNumberRecord(value: unknown): value is Record<string, number> {
  return isRecord(value) && Object.values(value).every(isNumber);
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

function isNumber(value: unknown): value is number {
  return typeof value === 'number' && Number.isFinite(value);
}

function isOptionalNumber(value: unknown): value is number | undefined {
  return value === undefined || isNumber(value);
}

function isOptionalString(value: unknown): value is string | undefined {
  return value === undefined || typeof value === 'string';
}
