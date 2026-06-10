import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState } from 'react';
import type { ReactNode } from 'react';
import {
  ApiError,
  approveJob,
  controlService,
  errorMessage,
  fetchAdminUsers,
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
  rejectJob,
} from './api';
import { AdminView } from './AdminView';
import type {
  AdminUser,
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
  { title: 'Govern', views: ['activity', 'admin', 'settings'] }
];

const validViews: View[] = navigationGroups.flatMap((group) => group.views);

type NavigationGroup = {
  title: string;
  views: View[];
};

type JobDecision = 'approve' | 'reject';
type ToastType = 'success' | 'error' | 'warning' | 'info';

export type ToastRequest = {
  type: ToastType;
  title: string;
  message: string;
};

type Toast = ToastRequest & {
  id: number;
  persistent: boolean;
};

type ToastContextValue = {
  notify: (toast: ToastRequest) => void;
};

const ToastContext = createContext<ToastContextValue | null>(null);
const toastAutoDismissMs = 5000;

function App() {
  return (
    <ToastProvider>
      <AppShell />
    </ToastProvider>
  );
}

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);
  const nextID = useRef(1);

  const dismiss = useCallback((id: number) => {
    setToasts((current) => current.filter((toast) => toast.id !== id));
  }, []);

  const notify = useCallback((request: ToastRequest) => {
    const toast = createToast(request, nextID.current);
    nextID.current += 1;
    setToasts((current) => [...current, toast].slice(-4));
    if (!toast.persistent) {
      window.setTimeout(() => dismiss(toast.id), toastAutoDismissMs);
    }
  }, [dismiss]);

  const value = useMemo(() => ({ notify }), [notify]);

  return (
    <ToastContext.Provider value={value}>
      {children}
      <ToastRegionContent toasts={toasts} onDismiss={dismiss} />
    </ToastContext.Provider>
  );
}

export function useToast(): ToastContextValue {
  const value = useContext(ToastContext);
  if (!value) {
    throw new Error('ToastProvider is required');
  }
  return value;
}

function createToast(request: ToastRequest, id: number): Toast {
  return {
    ...request,
    id,
    persistent: request.type === 'error' || request.type === 'warning'
  };
}

function AppShell() {
  const { notify } = useToast();
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
  const [adminUsers, setAdminUsers] = useState<AdminUser[]>([]);
  const [processSort, setProcessSort] = useState<ProcessSort>('memoryPct');
  const [apiError, setApiError] = useState<string | null>(null);
  const [activityError, setActivityError] = useState<string | null>(null);
  const [jobsError, setJobsError] = useState<string | null>(null);
  const [adminError, setAdminError] = useState<string | null>(null);
  const [serviceControlNotice, setServiceControlNotice] = useState<string | null>(null);
  const [overviewUpdatedAt, setOverviewUpdatedAt] = useState<Date | null>(null);
  const [jobsUpdatedAt, setJobsUpdatedAt] = useState<Date | null>(null);
  const [activityUpdatedAt, setActivityUpdatedAt] = useState<Date | null>(null);

  useHashNavigation(setCurrentView);
  usePolling('metrics', fetchMetrics, setMetrics, setApiError, 5000, true, setOverviewUpdatedAt, notify);
  usePolling('disk', fetchDisk, setDisk, setApiError, 30000, true, setOverviewUpdatedAt, notify);
  usePolling('health', fetchHealth, setHealth, setApiError, 15000, true, setOverviewUpdatedAt, notify);
  usePolling('readiness', fetchReadiness, setReadiness, setApiError, 15000, true, setOverviewUpdatedAt, notify);
  usePolling('observability metrics', fetchObservabilityMetrics, setObservability, setApiError, 15000, true, setOverviewUpdatedAt, notify);
  usePolling('services', fetchServices, setServices, setApiError, 15000, currentView === 'services', undefined, notify);
  usePolling('processes', fetchProcesses, setProcesses, setApiError, 30000, currentView === 'processes', undefined, notify);
  usePolling('jobs', fetchJobs, setJobs, setJobsError, 15000, currentView === 'jobs', setJobsUpdatedAt, notify);
  usePolling('activity', fetchActivity, setActivity, setActivityError, 60000, currentView === 'activity', setActivityUpdatedAt, notify);
  usePolling('admin users', fetchAdminUsers, setAdminUsers, setAdminError, 30000, currentView === 'admin', undefined, notify);

  const sortedProcesses = useMemo(
    () => sortedProcessList(processes, processSort),
    [processes, processSort]
  );

  async function requestServiceControl(service: string, action: ServiceControlAction) {
    try {
      const job = await controlService(service, action);
      setApiError(null);
      setServiceControlNotice(`Job ${job.jobId} was created. ${plainJobAction(action)} on ${service} is waiting for approval.`);
      notify({
        type: 'success',
        title: 'Job created',
        message: `Koji queued the ${action} request. It still needs approval before it can run.`
      });
      setServices(await fetchServices());
      setJobs(await fetchJobs());
    } catch (error: unknown) {
      const message = errorMessage(error, 'Service control transaction failed');
      setServiceControlNotice(message);
      notify({ type: 'error', title: 'Service control failed', message });
    }
  }

  async function requestJobDecision(job: JobRecord, decision: JobDecision, reason: string) {
    try {
      const decidedJob = decision === 'approve' ? await approveJob(job.id, reason) : await rejectJob(job.id, reason);
      setJobs((currentJobs) => replaceJob(currentJobs, decidedJob));
      setJobs(await fetchJobs());
      setJobsError(null);
      notify({
        type: 'success',
        title: decision === 'approve' ? 'Job approved' : 'Job rejected',
        message: decision === 'approve'
          ? 'Koji marked the job approved. The worker can advance it when the local agent is available.'
          : 'Koji rejected the job. It will not run.'
      });
    } catch (error: unknown) {
      const message = errorMessage(error, 'Job decision failed');
      setJobsError(message);
      notify({ type: 'error', title: 'Job decision failed', message });
    }
  }

  async function refreshAdminUsers() {
    setAdminUsers(await fetchAdminUsers());
  }

  return (
    <div className="app">
      <aside className="sidebar" aria-label="Primary navigation">
        <div className="brand">Koji</div>
        <nav className="nav-menu" aria-label="Koji sections">
          {navigationGroups.map((group) => (
            <div key={group.title} className="nav-group">
              <div className="nav-group-title">{group.title}</div>
              {group.views.map((view) => (
                <button
                  key={view}
                  className={`nav-item ${currentView === view ? 'active' : ''}`}
                  onClick={() => navigateTo(view)}
                  aria-current={currentView === view ? 'page' : undefined}
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

        <main className="main-content" tabIndex={-1}>
          {apiError && <ErrorBanner message={apiError} />}

          {currentView === 'overview' && (
            <Overview
              metrics={metrics}
              disk={disk}
              health={health}
              readiness={readiness}
              observability={observability}
              updatedAt={overviewUpdatedAt}
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
          {currentView === 'jobs' && (
            <JobsView jobs={jobs} error={jobsError} updatedAt={jobsUpdatedAt} onDecision={requestJobDecision} />
          )}
          {currentView === 'activity' && (
            <ActivityView events={activity} error={activityError} updatedAt={activityUpdatedAt} />
          )}
          {currentView === 'admin' && (
            <AdminView
              users={adminUsers}
              error={adminError}
              onRefresh={refreshAdminUsers}
              notify={notify}
              formatTimestamp={formatTimestamp}
            />
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
  enabled = true,
  setLastUpdated?: (date: Date) => void,
  notify?: (toast: ToastRequest) => void
) {
  const lastError = useRef<string | null>(null);

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
        setLastUpdated?.(new Date());
        lastError.current = null;
      } catch (error: unknown) {
        if (!isAbortError(error)) {
          const message = errorMessage(error, `Failed to fetch ${label}`);
          setApiError(message);
          const toast = pollingToast(error, label, message);
          if (toast && message !== lastError.current) {
            notify?.(toast);
            lastError.current = message;
          }
        }
      }
    }

    poll();
    const interval = window.setInterval(poll, intervalMs);
    return () => {
      controller.abort();
      window.clearInterval(interval);
    };
  }, [enabled, fetcher, intervalMs, label, notify, setApiError, setLastUpdated, setValue]);
}

function replaceJob(jobs: JobRecord[], updatedJob: JobRecord): JobRecord[] {
  return jobs.map((job) => (job.id === updatedJob.id ? updatedJob : job));
}

function Overview({
  metrics,
  disk,
  health,
  readiness,
  observability,
  updatedAt
}: {
  metrics: SystemMetrics | null;
  disk: DiskMetrics | null;
  health: HealthResponse | null;
  readiness: HealthResponse | null;
  observability: ObservabilityMetrics | null;
  updatedAt: Date | null;
}) {
  return (
    <div className="page-stack">
      <PanelMeta updatedAt={updatedAt} />
      <section className="dashboard-grid">
        <MetricCard
          title="CPU"
          value={metrics ? `${metrics.cpuUsage.toFixed(1)}%` : '—'}
          detail="Current host CPU utilization."
          tooltip="CPU usage shows current authorized host load. If it stays high, review running processes and queued jobs before approving more work."
        >
          <Gauge value={metrics?.cpuUsage ?? 0} tone="cpu" label="CPU" />
        </MetricCard>
        <MetricCard
          title="Memory"
          value={metrics ? `${formatGB(metrics.memUsed)} / ${formatGB(metrics.memTotal)} GB` : '—'}
          detail={metrics ? `${metrics.memUsagePct.toFixed(1)}% used` : 'Waiting for metrics'}
          tooltip="Memory usage compares used memory against total memory. If memory is tight, inspect process visibility and avoid approving disruptive jobs."
        >
          <Gauge value={metrics?.memUsagePct ?? 0} tone="memory" label="Memory" />
        </MetricCard>
        <MetricCard
          title="Disk"
          value={disk ? `${formatBytesAsGB(disk.usedBytes)} / ${formatBytesAsGB(disk.totalBytes)} GB` : '—'}
          detail={disk ? `${disk.usagePct.toFixed(1)}% used on ${disk.path}` : 'Waiting for disk metrics'}
          tooltip="Disk usage shows the configured filesystem. If it is nearly full, pause service changes until storage pressure is understood."
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
    <article className="metric-card" aria-label={title}>
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
    </article>
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
    <section className="status-panel" aria-label={`${title} status`}>
      <div className="status-panel-header">
        <span>
          {title}
          <Tooltip text={statusTooltip(title, response.status)} />
        </span>
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
    </section>
  );
}

export function ControlPlaneObservability({
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
        tooltip="Koji cannot run approved service jobs unless the local agent can be reached. Start or repair koji-agent if this shows agent unavailable."
      >
        <StatusBadge status={agentStatus(metrics, readiness)} label={plainAgentStatus(metrics, readiness)} />
      </MetricCard>
      <MetricCard
        title="Audit Writes"
        value={`${counter(metrics, 'audit_writes_total')} writes`}
        detail={`${counter(metrics, 'audit_write_failures_total')} failures`}
        tooltip="Audit writes show whether Koji is recording governed actions. If writes fail, pause approvals until the database is healthy."
      >
        <StatusBadge status={auditStatus(metrics)} label={plainAuditStatus(metrics)} />
      </MetricCard>
      <MetricCard
        title="Authentication"
        value={`${counter(metrics, 'auth_login_success_total')} successes`}
        detail={`${counter(metrics, 'auth_login_failure_total')} failed login attempts`}
        tooltip="Authentication metrics count login outcomes without exposing usernames. Repeated failures may indicate expired credentials or unauthorized access attempts."
      />
      <MetricCard
        title="Readiness Checks"
        value={`${counter(metrics, 'readiness_checks_total')} checks`}
        detail={readinessFailureDetail(metrics)}
        tooltip="Readiness checks summarize DB, migration, and agent dependency health. Failed DB or migration checks require operator attention before changes are approved."
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
        <InlineNotice
          tone="info"
          message={controlNotice}
          tooltip="Service control creates a durable job. It still requires approval, and the local agent must be reachable with mutation enabled before execution can happen."
        />
      )}
      <PermissionNotice message="Only services configured in the Koji allowlist are visible here. Other systemd units are hidden by policy." />
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
            <p className="help-text">Control buttons create approval-required jobs. They do not execute directly from the browser.</p>
            <div className="service-controls">
              <button type="button" onClick={() => onControl(service.name, service.active ? 'stop' : 'start')}>
                {service.active ? 'Stop' : 'Start'}
              </button>
              <button type="button" onClick={() => onControl(service.name, 'restart')}>Restart</button>
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
      <PermissionNotice message="Process details can be hidden by policy. Koji shows only the fields authorized by the backend." />
      <ProcessSummaryChart processes={processes} />
      <div className="process-container">
        <div className="table-controls">
          <span>
            Visible Processes: <strong>{processes.length}</strong>
            <Tooltip text="Some fields are hidden by process visibility policy. Use Settings to confirm policy intent; the UI cannot reveal redacted fields." />
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
                    {process.rss === undefined ? <HiddenByPolicy /> : `${formatMB(process.rss)} MB`}
                  </td>
                  <td className="font-tabular muted-cell" style={{ textAlign: 'right' }}>
                    {process.memoryPct === undefined ? <HiddenByPolicy /> : `${process.memoryPct.toFixed(1)}%`}
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

export function ActivityView({
  events,
  error,
  updatedAt
}: {
  events: ActivityEvent[];
  error: string | null;
  updatedAt: Date | null;
}) {
  if (error) {
    return <InlineError message={error} />;
  }
  if (events.length === 0) {
    return <EmptyState title="No activity available" detail="Audit activity appears here when your account has the read capability." />;
  }
  return (
    <div className="activity-panel">
      <div className="card-heading">
        <h3>Recent Activity</h3>
        <Tooltip text="Activity shows normalized audit events. Raw actor metadata, remote address, and internal error details are intentionally hidden." />
      </div>
      <PanelMeta updatedAt={updatedAt} />
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

export function JobsView({
  jobs,
  error,
  updatedAt,
  onDecision
}: {
  jobs: JobRecord[];
  error: string | null;
  updatedAt: Date | null;
  onDecision: (job: JobRecord, decision: JobDecision, reason: string) => Promise<void>;
}) {
  const [reasons, setReasons] = useState<Record<string, string>>({});

  if (error) {
    return <InlineError message={error} />;
  }
  if (jobs.length === 0) {
    return <EmptyState title="No jobs visible" detail="Service-control requests appear here after they are accepted as durable jobs." />;
  }
  return (
    <div className="activity-panel">
      <div className="card-heading">
        <h3>Jobs</h3>
        <Tooltip text="Jobs persist service-control intent. Jobs waiting for approval need a human decision before the worker can advance them." />
      </div>
      <PanelMeta updatedAt={updatedAt} />
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
      <span className="sr-only" id={`decision-help-${job.id}`}>
        Approval lets the worker advance this job when the agent is available. Rejection stops it from running.
      </span>
      <input
        aria-label={`Decision reason for ${job.id}`}
        aria-describedby={`decision-help-${job.id}`}
        maxLength={512}
        placeholder="Reason"
        value={reason}
        onChange={(event) => onReasonChange(event.target.value)}
      />
      <button type="button" onClick={() => onDecision(job, 'approve', reason)} aria-describedby={`decision-help-${job.id}`}>
        Approve
      </button>
      <button type="button" onClick={() => onDecision(job, 'reject', reason)} aria-describedby={`decision-help-${job.id}`}>
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
  return `${queued} waiting for approval, ${running} running, ${failed} failed`;
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
  return agentStatus(metrics, readiness) === 'ok' ? 'Agent reachable' : 'Agent unavailable';
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
  return `${dbFailures} DB failures, ${migrationFailures} migration failures, ${agentDegraded} agent unavailable events`;
}

function PanelMeta({ updatedAt }: { updatedAt: Date | null }) {
  return (
    <div className="panel-meta">
      <LastUpdated value={updatedAt} />
      <StaleDataNotice value={updatedAt} />
    </div>
  );
}

function LastUpdated({ value }: { value: Date | null }) {
  return <span className="last-updated">{value ? `Last updated ${formatTimeOnly(value)}` : 'Waiting for first update'}</span>;
}

function StaleDataNotice({ value }: { value: Date | null }) {
  if (!value || Date.now() - value.getTime() < 120000) {
    return null;
  }
  return <span className="stale-data-notice">Data may be stale. Check connection status before taking action.</span>;
}

function InlineError({ message }: { message: string }) {
  return <InlineNotice tone="error" message={message} />;
}

function PermissionNotice({ message }: { message: string }) {
  return <InlineNotice tone="info" message={message} />;
}

function InlineNotice({
  tone,
  message,
  tooltip
}: {
  tone: 'error' | 'info' | 'warning';
  message: string;
  tooltip?: string;
}) {
  return (
    <div className={`inline-notice ${tone}`} role={tone === 'error' ? 'alert' : 'status'}>
      <span>{noticePrefix(tone)} {message}</span>
      {tooltip && <Tooltip text={tooltip} />}
    </div>
  );
}

function HiddenByPolicy() {
  return <span className="policy-hidden">Hidden by policy</span>;
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

export type UIStatus = HealthStatus | 'running' | 'completed' | 'pending';

export function StatusBadge({ status, label }: { status: UIStatus; label?: string }) {
  return <span className={`status-badge ${status}`}>{statusIcon(status)} {label ?? statusLabel(status)}</span>;
}

export function ErrorBanner({ message, tooltip }: { message: string; tooltip?: string }) {
  return (
    <div className="error-banner" role="alert">
      <span>{message}</span>
      {tooltip && <Tooltip text={tooltip} />}
    </div>
  );
}

export function Tooltip({ text }: { text: string }) {
  const id = useTooltipID();
  return (
    <span className="tooltip" tabIndex={0} aria-label="Help" aria-describedby={id}>
      ?
      <span id={id} className="tooltip-content" role="tooltip">{text}</span>
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
  return <div className="loading-state" role="status" aria-live="polite">{label}</div>;
}

function ToastRegionContent({
  toasts,
  onDismiss
}: {
  toasts: Toast[];
  onDismiss: (id: number) => void;
}) {
  return (
    <div className="toast-region" aria-live="polite" aria-label="Notifications">
      {toasts.map((toast) => (
        <div key={toast.id} className={`toast ${toast.type}`} role={toast.type === 'error' ? 'alert' : 'status'}>
          <div>
            <strong>{toast.title}</strong>
            <p>{toast.message}</p>
          </div>
          <button type="button" onClick={() => onDismiss(toast.id)} aria-label={`Dismiss ${toast.title}`}>
            Dismiss
          </button>
        </div>
      ))}
    </div>
  );
}

function useTooltipID(): string {
  const idRef = useRef<string | null>(null);
  if (idRef.current === null) {
    idRef.current = `tooltip-${Math.random().toString(36).slice(2)}`;
  }
  return idRef.current;
}

function SettingsView() {
  return (
    <div className="settings-grid">
      <PolicyCard title="Session Policy" detail="Session lifetime and idle timeout are enforced by the backend. If a session expires, refresh and sign in again before continuing." />
      <PolicyCard title="Process Visibility" detail="Process fields are redacted unless the backend policy exposes them. Hidden fields cannot be recovered by the UI." />
      <PolicyCard title="Service Allowlist" detail="Only allowlisted systemd units are visible or eligible for control intent. Missing services require a configuration change." />
      <PolicyCard title="Capabilities" detail="Authenticated users still need explicit capabilities for protected views and job decisions." />
      <PolicyCard title="Agent Boundary" detail="The web daemon queues and approves intent. The local agent owns any future privileged mutation path." />
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
      return 'Requests waiting for approval, running, or completed.';
    case 'activity':
      return 'Governed audit activity read model.';
    case 'admin':
      return 'Users, capabilities, and one-time magic tokens.';
    case 'settings':
      return 'Read-only policy summary.';
  }
}

function toastTypeForMessage(message: string): ToastType {
  return message.includes('permission') || message.includes('Sign in') || message.includes('session')
    ? 'warning'
    : 'error';
}

function pollingToast(error: unknown, label: string, message: string): ToastRequest | null {
  if (!(error instanceof ApiError)) {
    return null;
  }
  if (!shouldToastPollingError(error)) {
    return null;
  }
  return {
    type: toastTypeForMessage(message),
    title: `${humanizeKey(label)} needs attention`,
    message
  };
}

function shouldToastPollingError(error: ApiError): boolean {
  return (
    error.code === 'network_error' ||
    error.code === 'unauthenticated' ||
    error.code === 'forbidden' ||
    error.code === 'csrf_missing_or_invalid' ||
    error.code === 'session_expired'
  );
}

function noticePrefix(tone: 'error' | 'info' | 'warning'): string {
  switch (tone) {
    case 'error':
      return 'Error:';
    case 'warning':
      return 'Warning:';
    case 'info':
      return 'Note:';
  }
}

function statusIcon(status: UIStatus): string {
  switch (status) {
    case 'ok':
      return 'OK';
    case 'degraded':
      return 'WARN';
    case 'fail':
      return 'FAIL';
    case 'running':
      return 'RUN';
    case 'completed':
      return 'DONE';
    case 'pending':
      return 'WAIT';
  }
}

function statusLabel(status: UIStatus): string {
  switch (status) {
    case 'ok':
      return 'Healthy';
    case 'degraded':
      return 'Degraded';
    case 'fail':
      return 'Failed';
    case 'running':
      return 'Running';
    case 'completed':
      return 'Completed';
    case 'pending':
      return 'Pending';
  }
}

function statusTooltip(title: string, status: HealthStatus): string {
  if (title === 'Readiness' && status === 'degraded') {
    return 'Koji is running, but one dependency is degraded. Check the agent before approving jobs.';
  }
  if (status === 'fail') {
    return `${title} failed. Pause operational changes and inspect the failing check.`;
  }
  return `${title} is healthy enough for normal operation.`;
}

function plainJobAction(action: string): string {
  return humanizeKey(action.replace(/\./g, ' '));
}

function plainJobStatus(status: JobStatus): string {
  switch (status) {
    case 'queued':
      return 'Waiting for approval';
    case 'approved':
      return 'Approved, waiting for worker';
    case 'rejected':
      return 'Rejected';
    case 'running':
      return 'Running';
    case 'completed':
      return 'Completed';
    case 'failed':
      return 'Failed';
    case 'not_implemented':
      return 'Agent not implemented';
  }
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
  switch (job.status) {
    case 'approved':
      return 'Approved. The worker can advance it when the agent is available.';
    case 'running':
      return 'Running through the worker.';
    case 'not_implemented':
      return 'The agent does not implement this action yet.';
    case 'failed':
      return 'Failed. Review the reason before creating a replacement job.';
    case 'completed':
      return 'Completed.';
    case 'rejected':
      return 'Rejected. It will not run.';
    case 'queued':
      return 'Waiting for approval.';
  }
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

function formatTimeOnly(value: Date): string {
  return value.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
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
