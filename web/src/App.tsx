import { useEffect, useMemo, useState } from 'react';
import type { ReactNode } from 'react';
import {
  approveJob,
  controlService,
  errorMessage,
  fetchActivity,
  fetchDisk,
  fetchHealth,
  fetchJobs,
  fetchMetrics,
  fetchObservabilityMetrics,
  fetchProcesses,
  fetchReadiness,
  fetchServices,
  isAbortError,
  rejectJob
} from './api';
import type {
  ActivityEvent,
  DiskMetrics,
  HealthResponse,
  HealthStatus,
  JobRecord,
  JobStatus,
  ObservabilityMetrics,
  ProcessInfo,
  ProcessSort,
  ServiceControlAction,
  ServiceStatus,
  SystemMetrics,
  View
} from './types';
import './App.css';

const navigationGroups: NavigationGroup[] = [
  { title: 'Operate', views: ['overview', 'services', 'processes', 'jobs'] },
  { title: 'Govern', views: ['activity', 'settings'] }
];

const validViews: View[] = navigationGroups.flatMap((group) => group.views);

type NavigationGroup = {
  title: string;
  views: View[];
};

type JobDecision = 'approve' | 'reject';

function App() {
  const [currentView, setCurrentView] = useState<View>(() => viewFromHash(window.location.hash));
  const [metrics, setMetrics] = useState<SystemMetrics | null>(null);
  const [disk, setDisk] = useState<DiskMetrics | null>(null);
  const [health, setHealth] = useState<HealthResponse | null>(null);
  const [readiness, setReadiness] = useState<HealthResponse | null>(null);
  const [observability, setObservability] = useState<ObservabilityMetrics | null>(null);
  const [services, setServices] = useState<ServiceStatus[]>([]);
  const [processes, setProcesses] = useState<ProcessInfo[]>([]);
  const [activity, setActivity] = useState<ActivityEvent[]>([]);
  const [jobs, setJobs] = useState<JobRecord[]>([]);
  const [processSort, setProcessSort] = useState<ProcessSort>('memoryPct');
  const [apiError, setApiError] = useState<string | null>(null);
  const [activityError, setActivityError] = useState<string | null>(null);
  const [jobsError, setJobsError] = useState<string | null>(null);
  const [serviceControlNotice, setServiceControlNotice] = useState<string | null>(null);

  useHashNavigation(setCurrentView);
  usePolling('metrics', fetchMetrics, setMetrics, setApiError, 2000);
  usePolling('disk', fetchDisk, setDisk, setApiError, 10000);
  usePolling('health', fetchHealth, setHealth, setApiError, 15000);
  usePolling('readiness', fetchReadiness, setReadiness, setApiError, 15000);
  usePolling('observability metrics', fetchObservabilityMetrics, setObservability, setApiError, 7000);
  usePolling('services', fetchServices, setServices, setApiError, 5000);
  usePolling('processes', fetchProcesses, setProcesses, setApiError, 4000);
  usePolling('jobs', fetchJobs, setJobs, setJobsError, 7000, currentView === 'jobs');
  usePolling('activity', fetchActivity, setActivity, setActivityError, 10000, currentView === 'activity');

  const sortedProcesses = useMemo(
    () => sortedProcessList(processes, processSort),
    [processes, processSort]
  );

  async function requestServiceControl(service: string, action: ServiceControlAction) {
    try {
      const job = await controlService(service, action);
      setApiError(null);
      setServiceControlNotice(`Job ${job.jobId} queued for ${action} on ${service}.`);
      setServices(await fetchServices());
      setJobs(await fetchJobs());
    } catch (error: unknown) {
      setServiceControlNotice(errorMessage(error, 'Service control transaction failed'));
    }
  }

  async function requestJobDecision(job: JobRecord, decision: JobDecision, reason: string) {
    try {
      const decidedJob = decision === 'approve' ? await approveJob(job.id, reason) : await rejectJob(job.id, reason);
      setJobs((currentJobs) => replaceJob(currentJobs, decidedJob));
      setJobs(await fetchJobs());
      setJobsError(null);
    } catch (error: unknown) {
      setJobsError(errorMessage(error, 'Job decision failed'));
    }
  }

  return (
    <div className="app">
      <aside className="sidebar">
        <div className="brand">Koji</div>
        <nav className="nav-menu">
          {navigationGroups.map((group) => (
            <div key={group.title} className="nav-group">
              <div className="nav-group-title">{group.title}</div>
              {group.views.map((view) => (
                <button
                  key={view}
                  className={`nav-item ${currentView === view ? 'active' : ''}`}
                  onClick={() => navigateTo(view)}
                >
                  {viewLabel(view)}
                </button>
              ))}
            </div>
          ))}
        </nav>
        <div className="sidebar-footer">
          <ConnectionStatus apiError={apiError} metrics={metrics} />
        </div>
      </aside>

      <div className="content-area">
        <header className="top-bar">
          <div>
            <div className="view-title">{viewLabel(currentView)}</div>
            <div className="view-subtitle">{viewSubtitle(currentView)}</div>
          </div>
          {metrics && <div className="quick-stats">Uptime: {formatUptime(metrics.uptime)}</div>}
        </header>

        <main className="main-content">
          {apiError && <ErrorBanner message={apiError} />}

          {currentView === 'overview' && (
            <Overview
              metrics={metrics}
              disk={disk}
              health={health}
              readiness={readiness}
              observability={observability}
            />
          )}
          {currentView === 'services' && (
            <ServicesView
              services={services}
              controlNotice={serviceControlNotice}
              onControl={requestServiceControl}
            />
          )}
          {currentView === 'processes' && (
            <ProcessesView
              processes={processes}
              sortedProcesses={sortedProcesses}
              processSort={processSort}
              onSortChange={setProcessSort}
            />
          )}
          {currentView === 'jobs' && <JobsView jobs={jobs} error={jobsError} onDecision={requestJobDecision} />}
          {currentView === 'activity' && (
            <ActivityView events={activity} error={activityError} />
          )}
          {currentView === 'settings' && <SettingsView />}
        </main>
      </div>
    </div>
  );
}

function useHashNavigation(setCurrentView: (view: View) => void) {
  useEffect(() => {
    const handleHashChange = () => {
      setCurrentView(viewFromHash(window.location.hash));
    };
    window.addEventListener('hashchange', handleHashChange);
    return () => window.removeEventListener('hashchange', handleHashChange);
  }, [setCurrentView]);
}

function usePolling<T>(
  label: string,
  fetcher: (signal?: AbortSignal) => Promise<T>,
  setValue: (value: T) => void,
  setApiError: (message: string | null) => void,
  intervalMs: number,
  enabled = true
) {
  useEffect(() => {
    if (!enabled) {
      return;
    }
    const controller = new AbortController();

    async function poll() {
      try {
        const value = await fetcher(controller.signal);
        setValue(value);
        setApiError(null);
      } catch (error: unknown) {
        if (!isAbortError(error)) {
          setApiError(errorMessage(error, `Failed to fetch ${label}`));
        }
      }
    }

    poll();
    const interval = window.setInterval(poll, intervalMs);
    return () => {
      controller.abort();
      window.clearInterval(interval);
    };
  }, [enabled, fetcher, intervalMs, label, setApiError, setValue]);
}

function replaceJob(jobs: JobRecord[], updatedJob: JobRecord): JobRecord[] {
  return jobs.map((job) => (job.id === updatedJob.id ? updatedJob : job));
}

function Overview({
  metrics,
  disk,
  health,
  readiness,
  observability
}: {
  metrics: SystemMetrics | null;
  disk: DiskMetrics | null;
  health: HealthResponse | null;
  readiness: HealthResponse | null;
  observability: ObservabilityMetrics | null;
}) {
  return (
    <div className="page-stack">
      <section className="dashboard-grid">
        <MetricCard
          title="CPU"
          value={metrics ? `${metrics.cpuUsage.toFixed(1)}%` : '—'}
          detail="Current host CPU utilization."
          tooltip="CPU percentage is already-authorized telemetry from the Koji metrics API."
        >
          <Gauge value={metrics?.cpuUsage ?? 0} tone="cpu" label="CPU" />
        </MetricCard>
        <MetricCard
          title="Memory"
          value={metrics ? `${formatGB(metrics.memUsed)} / ${formatGB(metrics.memTotal)} GB` : '—'}
          detail={metrics ? `${metrics.memUsagePct.toFixed(1)}% used` : 'Waiting for metrics'}
          tooltip="Memory usage compares used memory against total memory reported by the host."
        >
          <Gauge value={metrics?.memUsagePct ?? 0} tone="memory" label="Memory" />
        </MetricCard>
        <MetricCard
          title="Disk"
          value={disk ? `${formatBytesAsGB(disk.usedBytes)} / ${formatBytesAsGB(disk.totalBytes)} GB` : '—'}
          detail={disk ? `${disk.usagePct.toFixed(1)}% used on ${disk.path}` : 'Waiting for disk metrics'}
          tooltip="Disk usage is reported for the filesystem exposed by the authorized disk API."
        >
          <Gauge value={disk?.usagePct ?? 0} tone="filesystem" label="Disk" />
        </MetricCard>
        <MetricCard
          title="Uptime"
          value={metrics ? formatUptime(metrics.uptime) : '—'}
          detail="Time since the host last booted."
        />
      </section>
      <section className="status-grid">
        <OperationalStatus title="Health" response={health} />
        <OperationalStatus title="Readiness" response={readiness} />
      </section>
      <ControlPlaneObservability metrics={observability} readiness={readiness} />
    </div>
  );
}

function MetricCard({
  title,
  value,
  detail,
  tooltip,
  children
}: {
  title: string;
  value: string;
  detail: string;
  tooltip?: string;
  children?: ReactNode;
}) {
  return (
    <div className="metric-card">
      <div className="card-heading">
        <h3>{title}</h3>
        {tooltip && <Tooltip text={tooltip} />}
      </div>
      <div className="metric-body">
        {children}
        <div>
          <div className="metric-value">{value}</div>
          <small>{detail}</small>
        </div>
      </div>
    </div>
  );
}

function Gauge({ value, tone, label }: { value: number; tone: string; label: string }) {
  const percent = boundedPercent(value);
  return (
    <div className={`gauge ${tone}`} aria-label={`${label} ${percent.toFixed(0)} percent`}>
      <div className="gauge-fill" style={{ transform: `rotate(${gaugeRotation(percent)}deg)` }} />
      <div className="gauge-mask" />
      <div className="gauge-value">{percent.toFixed(0)}%</div>
    </div>
  );
}

function OperationalStatus({ title, response }: { title: string; response: HealthResponse | null }) {
  if (!response) {
    return <LoadingState label={`${title} check pending`} />;
  }
  return (
    <div className="status-panel">
      <div className="status-panel-header">
        <span>{title}</span>
        <StatusBadge status={response.status} />
      </div>
      <div className="check-list">
        {Object.entries(response.checks).map(([name, check]) => (
          <div key={name} className="check-row">
            <span>{humanizeKey(name)}</span>
            <StatusBadge status={check.status} />
          </div>
        ))}
      </div>
    </div>
  );
}

function ControlPlaneObservability({
  metrics,
  readiness
}: {
  metrics: ObservabilityMetrics | null;
  readiness: HealthResponse | null;
}) {
  if (!metrics) {
    return <LoadingState label="Loading control-plane metrics" />;
  }
  return (
    <section className="control-plane-grid">
      <MetricCard
        title="Job Flow"
        value={`${counter(metrics, 'jobs_created_total')} created`}
        detail={jobFlowDetail(metrics)}
        tooltip="Job counters track durable service-control intent and approval lifecycle events."
      />
      <MetricCard
        title="Worker"
        value={`${counter(metrics, 'worker_polls_total')} polls`}
        detail={`${counter(metrics, 'worker_errors_total')} worker errors`}
        tooltip="Worker metrics show whether approved jobs are being polled and advanced by kojid."
      >
        <StatusBadge status={workerStatus(metrics)} label={plainWorkerStatus(metrics)} />
      </MetricCard>
      <MetricCard
        title="Agent RPC"
        value={`${counter(metrics, 'agent_rpc_requests_total')} requests`}
        detail={`${counter(metrics, 'agent_rpc_failures_total')} failures`}
        tooltip="Agent RPC metrics cover kojid calls across the Unix socket boundary."
      >
        <StatusBadge status={agentStatus(metrics, readiness)} label={plainAgentStatus(metrics, readiness)} />
      </MetricCard>
      <MetricCard
        title="Audit Writes"
        value={`${counter(metrics, 'audit_writes_total')} writes`}
        detail={`${counter(metrics, 'audit_write_failures_total')} failures`}
        tooltip="Audit metrics show whether governance events are being persisted."
      >
        <StatusBadge status={auditStatus(metrics)} label={plainAuditStatus(metrics)} />
      </MetricCard>
      <MetricCard
        title="Authentication"
        value={`${counter(metrics, 'auth_login_success_total')} successes`}
        detail={`${counter(metrics, 'auth_login_failure_total')} failed login attempts`}
        tooltip="Authentication metrics count login outcomes without exposing usernames or session data."
      />
      <MetricCard
        title="Readiness Checks"
        value={`${counter(metrics, 'readiness_checks_total')} checks`}
        detail={readinessFailureDetail(metrics)}
        tooltip="Readiness counters summarize dependency failures without exposing sensitive connection details."
      />
    </section>
  );
}

function ServicesView({
  services,
  controlNotice,
  onControl
}: {
  services: ServiceStatus[];
  controlNotice: string | null;
  onControl: (service: string, action: ServiceControlAction) => void;
}) {
  return (
    <div className="page-stack">
      {controlNotice && (
        <ErrorBanner
          message={controlNotice}
          tooltip="Service control depends on the local Koji agent and remains unavailable when the agent is down or the build returns not implemented."
        />
      )}
      <div className="services-grid">
        {services.length === 0 && (
          <EmptyState title="No allowlisted services visible" detail="Only services configured in the backend allowlist appear here." />
        )}
        {services.map((service) => (
          <div key={service.name} className="service-card">
            <div className="service-header">
              <span className="service-name">{service.name.replace('.service', '')}</span>
              <StatusBadge status={service.active ? 'ok' : 'degraded'} label={service.active ? 'active' : 'inactive'} />
            </div>
            <div className="service-state">{service.subState}</div>
            <div className="service-controls">
              <button onClick={() => onControl(service.name, service.active ? 'stop' : 'start')}>
                {service.active ? 'Stop' : 'Start'}
              </button>
              <button onClick={() => onControl(service.name, 'restart')}>Restart</button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function ProcessesView({
  processes,
  sortedProcesses,
  processSort,
  onSortChange
}: {
  processes: ProcessInfo[];
  sortedProcesses: ProcessInfo[];
  processSort: ProcessSort;
  onSortChange: (sort: ProcessSort) => void;
}) {
  return (
    <div className="page-stack">
      <ProcessSummaryChart processes={processes} />
      <div className="process-container">
        <div className="table-controls">
          <span>
            Visible Processes: <strong>{processes.length}</strong>
            <Tooltip text="Some fields are redacted by backend process visibility policy. The UI does not reconstruct or request hidden data." />
          </span>
          <div className="toggle-group">
            <button
              className={processSort === 'memoryPct' ? 'active' : ''}
              onClick={() => onSortChange('memoryPct')}
            >
              Sort by Memory
            </button>
            <button className={processSort === 'pid' ? 'active' : ''} onClick={() => onSortChange('pid')}>
              Sort by PID
            </button>
          </div>
        </div>
        <div className="table-wrapper">
          <table className="process-table">
            <thead>
              <tr>
                <th style={{ width: '80px' }}>PID</th>
                <th>Name</th>
                <th style={{ width: '80px' }}>State</th>
                <th style={{ width: '120px', textAlign: 'right' }}>
                  RSS <Tooltip text="RSS is hidden unless the process visibility policy allows detailed fields." />
                </th>
                <th style={{ width: '100px', textAlign: 'right' }}>% Mem</th>
              </tr>
            </thead>
            <tbody>
              {sortedProcesses.length === 0 && (
                <tr>
                  <td colSpan={5} className="table-empty">
                    <LoadingState label="Reading authorized process view" />
                  </td>
                </tr>
              )}
              {sortedProcesses.map((process) => (
                <tr key={process.pid}>
                  <td className="font-tabular">{process.pid}</td>
                  <td className="proc-name" title={process.commandLine ?? process.name}>
                    {process.commandLine ?? process.name}
                  </td>
                  <td>
                    <span className={`state-badge ${process.state}`}>{process.state}</span>
                  </td>
                  <td className="font-tabular muted-cell" style={{ textAlign: 'right' }}>
                    {process.rss === undefined ? 'redacted' : `${formatMB(process.rss)} MB`}
                  </td>
                  <td className="font-tabular muted-cell" style={{ textAlign: 'right' }}>
                    {process.memoryPct === undefined ? 'redacted' : `${process.memoryPct.toFixed(1)}%`}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

function ProcessSummaryChart({ processes }: { processes: ProcessInfo[] }) {
  const buckets = processStateBuckets(processes);
  return (
    <div className="chart-panel">
      <div className="card-heading">
        <h3>Process States</h3>
        <Tooltip text="This chart uses only PID, name, and state, which remain available in summary mode." />
      </div>
      <div className="bar-chart">
        {buckets.map((bucket) => (
          <div key={bucket.state} className="bar-row">
            <span>{bucket.state}</span>
            <div className="bar-track">
              <div className="bar-fill" style={{ width: `${bucket.percent}%` }} />
            </div>
            <strong>{bucket.count}</strong>
          </div>
        ))}
      </div>
    </div>
  );
}

function ActivityView({ events, error }: { events: ActivityEvent[]; error: string | null }) {
  if (error) {
    return <ErrorBanner message={error} />;
  }
  if (events.length === 0) {
    return <EmptyState title="No activity available" detail="Audit activity appears here when your account has the read capability." />;
  }
  return (
    <div className="activity-panel">
      <div className="card-heading">
        <h3>Recent Activity</h3>
        <Tooltip text="Activity is a normalized audit read model. Raw actor, remote address, and internal error details are not exposed." />
      </div>
      <div className="table-wrapper">
        <table className="activity-table">
          <thead>
            <tr>
              <th>Time</th>
              <th>Action</th>
              <th>Target</th>
              <th>Outcome</th>
              <th>Reason</th>
              <th>Request ID</th>
            </tr>
          </thead>
          <tbody>
            {events.map((event) => (
              <tr key={`${event.timestamp}-${event.request_id}-${event.action}`}>
                <td className="font-tabular">{formatTimestamp(event.timestamp)}</td>
                <td>{plainAction(event.action)}</td>
                <td>{event.target}</td>
                <td>
                  <StatusBadge status={activityStatus(event.outcome)} label={plainOutcome(event.outcome)} />
                </td>
                <td>{plainReason(event.reason_code)}</td>
                <td className="request-id" title={event.request_id}>
                  {event.request_id || 'none'}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function JobsView({
  jobs,
  error,
  onDecision
}: {
  jobs: JobRecord[];
  error: string | null;
  onDecision: (job: JobRecord, decision: JobDecision, reason: string) => Promise<void>;
}) {
  const [reasons, setReasons] = useState<Record<string, string>>({});

  if (error) {
    return <ErrorBanner message={error} />;
  }
  if (jobs.length === 0) {
    return <EmptyState title="No jobs visible" detail="Service-control requests appear here after they are accepted as durable jobs." />;
  }
  return (
    <div className="activity-panel">
      <div className="card-heading">
        <h3>Jobs</h3>
        <Tooltip text="Jobs persist service-control intent before any privileged execution is enabled." />
      </div>
      <div className="table-wrapper">
        <table className="activity-table">
          <thead>
            <tr>
              <th>Created</th>
              <th>Action</th>
              <th>Target</th>
              <th>Status</th>
              <th>Reason</th>
              <th>Request ID</th>
              <th>Decision</th>
            </tr>
          </thead>
          <tbody>
            {jobs.map((job) => (
              <tr key={job.id}>
                <td className="font-tabular">{formatTimestamp(job.created_at)}</td>
                <td>{plainJobAction(job.action)}</td>
                <td>{job.target}</td>
                <td>
                  <StatusBadge status={jobStatusTone(job.status)} label={plainJobStatus(job.status)} />
                </td>
                <td>{plainReason(job.status_reason)}</td>
                <td className="request-id" title={job.request_id}>
                  {job.request_id || 'none'}
                </td>
                <td>
                  <JobDecisionControls
                    job={job}
                    reason={reasons[job.id] ?? ''}
                    onReasonChange={(reason) => setReasons((current) => ({ ...current, [job.id]: reason }))}
                    onDecision={onDecision}
                  />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function JobDecisionControls({
  job,
  reason,
  onReasonChange,
  onDecision
}: {
  job: JobRecord;
  reason: string;
  onReasonChange: (reason: string) => void;
  onDecision: (job: JobRecord, decision: JobDecision, reason: string) => Promise<void>;
}) {
  if (job.status !== 'queued') {
    return <span className="muted-text">{decisionSummary(job)}</span>;
  }
  return (
    <div className="job-decision-controls">
      <input
        aria-label={`Decision reason for ${job.id}`}
        maxLength={512}
        placeholder="Reason"
        value={reason}
        onChange={(event) => onReasonChange(event.target.value)}
      />
      <button type="button" onClick={() => onDecision(job, 'approve', reason)}>
        Approve
      </button>
      <button type="button" onClick={() => onDecision(job, 'reject', reason)}>
        Reject
      </button>
    </div>
  );
}

function counter(metrics: ObservabilityMetrics, name: string): number {
  return metrics.counters[name] ?? 0;
}

function statusCount(metrics: ObservabilityMetrics, status: JobStatus): number {
  return metrics.jobs_by_status[status] ?? 0;
}

function jobFlowDetail(metrics: ObservabilityMetrics): string {
  const queued = statusCount(metrics, 'queued');
  const running = statusCount(metrics, 'running');
  const failed = statusCount(metrics, 'failed');
  return `${queued} queued, ${running} running, ${failed} failed`;
}

function workerStatus(metrics: ObservabilityMetrics): HealthStatus {
  return counter(metrics, 'worker_errors_total') === 0 ? 'ok' : 'degraded';
}

function plainWorkerStatus(metrics: ObservabilityMetrics): string {
  return workerStatus(metrics) === 'ok' ? 'polling' : 'errors seen';
}

function agentStatus(metrics: ObservabilityMetrics, readiness: HealthResponse | null): HealthStatus {
  if (readiness?.checks.agent?.status === 'degraded') {
    return 'degraded';
  }
  return counter(metrics, 'agent_rpc_failures_total') === 0 ? 'ok' : 'degraded';
}

function plainAgentStatus(metrics: ObservabilityMetrics, readiness: HealthResponse | null): string {
  return agentStatus(metrics, readiness) === 'ok' ? 'reachable' : 'degraded';
}

function auditStatus(metrics: ObservabilityMetrics): HealthStatus {
  return counter(metrics, 'audit_write_failures_total') === 0 ? 'ok' : 'fail';
}

function plainAuditStatus(metrics: ObservabilityMetrics): string {
  return auditStatus(metrics) === 'ok' ? 'persisting' : 'write failures';
}

function readinessFailureDetail(metrics: ObservabilityMetrics): string {
  const dbFailures = counter(metrics, 'readiness_db_failures_total');
  const migrationFailures = counter(metrics, 'readiness_migration_failures_total');
  const agentDegraded = counter(metrics, 'readiness_agent_degraded_total');
  return `${dbFailures} DB failures, ${migrationFailures} migration failures, ${agentDegraded} agent degraded`;
}

function ConnectionStatus({ apiError, metrics }: { apiError: string | null; metrics: SystemMetrics | null }) {
  if (apiError) {
    return <span className="status error">● Disconnected</span>;
  }
  if (metrics) {
    return <span className="status online">● Online</span>;
  }
  return <span className="status">Connecting...</span>;
}

function StatusBadge({ status, label }: { status: HealthStatus; label?: string }) {
  return <span className={`status-badge ${status}`}>{label ?? status}</span>;
}

function ErrorBanner({ message, tooltip }: { message: string; tooltip?: string }) {
  return (
    <div className="error-banner">
      <span>{message}</span>
      {tooltip && <Tooltip text={tooltip} />}
    </div>
  );
}

function Tooltip({ text }: { text: string }) {
  return (
    <span className="tooltip" tabIndex={0} aria-label={text}>
      ?
      <span className="tooltip-content">{text}</span>
    </span>
  );
}

function EmptyState({ title, detail }: { title: string; detail: string }) {
  return (
    <div className="empty-state">
      <strong>{title}</strong>
      <span>{detail}</span>
    </div>
  );
}

function LoadingState({ label }: { label: string }) {
  return <div className="loading-state">{label}</div>;
}

function SettingsView() {
  return (
    <div className="settings-grid">
      <PolicyCard title="Session Policy" detail="Session lifetime and idle timeout are enforced by the backend." />
      <PolicyCard title="Process Visibility" detail="Process fields are redacted unless the backend policy exposes them." />
      <PolicyCard title="Service Allowlist" detail="Only allowlisted systemd units are visible or eligible for control intent." />
    </div>
  );
}

function PolicyCard({ title, detail }: { title: string; detail: string }) {
  return (
    <div className="policy-card">
      <h3>{title}</h3>
      <p>{detail}</p>
    </div>
  );
}

function viewFromHash(hash: string): View {
  const candidate = hash.replace('#', '');
  return isView(candidate) ? candidate : 'overview';
}

function isView(value: string): value is View {
  return validViews.includes(value as View);
}

function navigateTo(view: View) {
  window.location.hash = `#${view}`;
}

function viewLabel(view: View): string {
  return view.charAt(0).toUpperCase() + view.slice(1);
}

function viewSubtitle(view: View): string {
  switch (view) {
    case 'overview':
      return 'Safe host summary and operational readiness.';
    case 'services':
      return 'Allowlisted service state and control intent.';
    case 'processes':
      return 'Redaction-aware process visibility.';
    case 'jobs':
      return 'Durable service-control job lifecycle.';
    case 'activity':
      return 'Governed audit activity read model.';
    case 'settings':
      return 'Read-only policy summary.';
  }
}

function plainJobAction(action: string): string {
  return humanizeKey(action.replace(/\./g, ' '));
}

function plainJobStatus(status: JobStatus): string {
  return humanizeKey(status);
}

function jobStatusTone(status: JobStatus): HealthStatus {
  switch (status) {
    case 'completed':
    case 'approved':
      return 'ok';
    case 'queued':
    case 'running':
    case 'not_implemented':
      return 'degraded';
    case 'rejected':
    case 'failed':
      return 'fail';
  }
}

function decisionSummary(job: JobRecord): string {
  if (job.decision_reason !== '') {
    return plainReason(job.decision_reason);
  }
  return plainJobStatus(job.status);
}

function plainAction(action: string): string {
  const labels: Record<string, string> = {
    'auth.bootstrap': 'Bootstrap',
    'auth.login': 'Login',
    'auth.logout': 'Logout',
    'capability.denied': 'Capability denied',
    'capability.bypass': 'Dev bypass',
    'service.control': 'Service control intent',
    'job.approved': 'Job approved',
    'job.rejected': 'Job rejected',
    'job.approval_denied': 'Job approval denied',
    'process.list': 'Process list read'
  };
  return labels[action] ?? humanizeKey(action.replace(/\./g, ' '));
}

function plainOutcome(outcome: string): string {
  return humanizeKey(outcome);
}

function plainReason(reason: string): string {
  if (reason === '') {
    return 'none';
  }
  return humanizeKey(reason);
}

function activityStatus(outcome: string): HealthStatus {
  switch (outcome) {
    case 'success':
    case 'accepted':
      return 'ok';
    case 'denied':
    case 'rejected':
      return 'degraded';
    default:
      return 'fail';
  }
}

function formatTimestamp(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleString();
}

function sortedProcessList(processes: ProcessInfo[], processSort: ProcessSort): ProcessInfo[] {
  return [...processes].sort((left, right) => {
    if (processSort === 'pid') {
      return left.pid - right.pid;
    }
    return (right.memoryPct ?? 0) - (left.memoryPct ?? 0);
  });
}

function processStateBuckets(processes: ProcessInfo[]): ProcessStateBucket[] {
  const counts = new Map<string, number>();
  for (const process of processes) {
    counts.set(process.state, (counts.get(process.state) ?? 0) + 1);
  }
  const total = Math.max(processes.length, 1);
  return Array.from(counts.entries())
    .sort(([left], [right]) => left.localeCompare(right))
    .map(([state, count]) => ({
      state,
      count,
      percent: (count / total) * 100
    }));
}

type ProcessStateBucket = {
  state: string;
  count: number;
  percent: number;
};

function formatGB(kb: number): string {
  return (kb / 1024 / 1024).toFixed(2);
}

function formatMB(bytes: number): string {
  return (bytes / 1024 / 1024).toFixed(1);
}

function formatBytesAsGB(bytes: number): string {
  return (bytes / 1024 / 1024 / 1024).toFixed(1);
}

function formatUptime(seconds: number): string {
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  return `${days}d ${hours}h ${minutes}m`;
}

function boundedPercent(value: number): number {
  return Math.max(0, Math.min(100, value));
}

function gaugeRotation(value: number): number {
  return (boundedPercent(value) / 100) * 180;
}

function humanizeKey(value: string): string {
  return value.replace(/_/g, ' ').replace(/\b\w/g, (letter) => letter.toUpperCase());
}

export default App;
