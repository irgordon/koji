export type View = 'overview' | 'services' | 'processes' | 'jobs' | 'activity' | 'settings';

export type ProcessSort = 'pid' | 'memoryPct';

export type ServiceControlAction = 'start' | 'stop' | 'restart';

export interface SystemMetrics {
  cpuUsage: number;
  memTotal: number;
  memAvailable: number;
  memUsed: number;
  memUsagePct: number;
  uptime: number;
}

export interface DiskMetrics {
  path: string;
  totalBytes: number;
  freeBytes: number;
  usedBytes: number;
  usagePct: number;
}

export interface ServiceStatus {
  name: string;
  active: boolean;
  subState: string;
}

export interface ServiceListResponse {
  services: ServiceStatus[];
}

export type HealthStatus = 'ok' | 'degraded' | 'fail';

export interface HealthCheck {
  status: HealthStatus;
}

export interface HealthResponse {
  status: HealthStatus;
  checks: Record<string, HealthCheck>;
}

export interface ProcessInfo {
  pid: number;
  name: string;
  state: string;
  uid?: number;
  ppid?: number;
  cpuUser?: number;
  cpuSystem?: number;
  rss?: number;
  memoryPct?: number;
  commandLine?: string;
}

export interface ActivityEvent {
  timestamp: string;
  action: string;
  target: string;
  outcome: string;
  reason_code: string;
  request_id: string;
}

export interface ActivityResponse {
  events: ActivityEvent[];
}

export type JobStatus =
  | 'queued'
  | 'approved'
  | 'rejected'
  | 'running'
  | 'completed'
  | 'failed'
  | 'not_implemented';

export interface JobRecord {
  id: string;
  created_at: string;
  created_by?: number;
  action: string;
  target: string;
  status: JobStatus;
  status_reason: string;
  request_id: string;
  approved_by?: number;
  approved_at?: string;
  rejected_by?: number;
  rejected_at?: string;
  decision_reason: string;
  started_at?: string;
}

export interface JobsResponse {
  jobs: JobRecord[];
}

export interface ObservabilityMetrics {
  counters: Record<string, number>;
  jobs_by_status: Record<string, number>;
}

export interface ServiceControlJobResponse {
  jobId: string;
  status: JobStatus;
}

export interface ApiErrorResponse {
  error: string;
}

export interface ApiStatusResponse {
  status: string;
}

export type NormalizedErrorCode =
  | 'unauthenticated'
  | 'forbidden'
  | 'csrf_missing_or_invalid'
  | 'agent_unavailable'
  | 'agent_not_implemented'
  | 'mutation_disabled'
  | 'service_not_allowlisted'
  | 'job_conflict'
  | 'validation_error'
  | 'network_error'
  | 'unexpected_response'
  | 'session_expired';
