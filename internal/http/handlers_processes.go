package http

import (
	"net/http"

	"koji/internal/audit"
	"koji/internal/caps"
	"koji/internal/config"
	"koji/internal/system"
)

func (h protectedHandlers) handleProcessesList(w http.ResponseWriter, r *http.Request) {
	h.handleProcessesListWithLister(system.ListProcesses)(w, r)
}

func (h protectedHandlers) handleProcessesListWithLister(lister processLister) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.serveProcessesList(w, r, lister)
	}
}

func (h protectedHandlers) serveProcessesList(w http.ResponseWriter, r *http.Request, lister processLister) {
	if !h.requireCapability(w, r, caps.HostProcessesRead, "host.processes") {
		return
	}

	processes, err := lister()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to list processes")
		return
	}

	if !h.recordProcessList(w, r) {
		return
	}

	writeJSONValue(w, http.StatusOK, h.processPolicy.apply(processes))
}

type processLister func() ([]system.ProcessInfo, error)

type processVisibilityPolicy struct {
	mode               string
	includeCommandLine bool
	maxProcesses       int
}

type processView struct {
	PID         int      `json:"pid"`
	Name        string   `json:"name"`
	State       string   `json:"state"`
	UID         *int     `json:"uid,omitempty"`
	PPID        *int     `json:"ppid,omitempty"`
	CPUUser     *uint64  `json:"cpuUser,omitempty"`
	CPUSystem   *uint64  `json:"cpuSystem,omitempty"`
	RSS         *int64   `json:"rss,omitempty"`
	MemoryPct   *float64 `json:"memoryPct,omitempty"`
	CommandLine string   `json:"commandLine,omitempty"`
}

func newProcessVisibilityPolicy(cfg config.Config) processVisibilityPolicy {
	return processVisibilityPolicy{
		mode:               cfg.ProcessVisibilityMode,
		includeCommandLine: cfg.IncludeCommandLine,
		maxProcesses:       cfg.MaxProcesses,
	}
}

func (policy processVisibilityPolicy) apply(processes []system.ProcessInfo) []processView {
	limit := policy.processLimit(len(processes))
	views := make([]processView, 0, limit)

	for _, process := range processes[:limit] {
		views = append(views, policy.view(process))
	}
	return views
}

func (policy processVisibilityPolicy) processLimit(count int) int {
	if policy.maxProcesses <= 0 || policy.maxProcesses > count {
		return count
	}
	return policy.maxProcesses
}

func (policy processVisibilityPolicy) view(process system.ProcessInfo) processView {
	view := processView{
		PID:   process.PID,
		Name:  process.Name,
		State: process.State,
	}
	policy.applyOwnerFields(&view, process)
	policy.applyDetailFields(&view, process)
	policy.applyCommandLine(&view, process)
	return view
}

func (policy processVisibilityPolicy) applyOwnerFields(view *processView, process system.ProcessInfo) {
	if policy.mode == config.ProcessVisibilityOwner || policy.mode == config.ProcessVisibilityAll {
		view.UID = &process.UID
	}
}

func (policy processVisibilityPolicy) applyDetailFields(view *processView, process system.ProcessInfo) {
	if policy.mode != config.ProcessVisibilityAll {
		return
	}
	view.PPID = &process.PPID
	view.CPUUser = &process.CPUUser
	view.CPUSystem = &process.CPUSystem
	view.RSS = &process.RSS
	view.MemoryPct = &process.MemoryPct
}

func (policy processVisibilityPolicy) applyCommandLine(view *processView, process system.ProcessInfo) {
	if policy.includeCommandLine {
		view.CommandLine = process.CommandLine
	}
}

func (h protectedHandlers) recordProcessList(w http.ResponseWriter, r *http.Request) bool {
	principal, ok := principalFromRequest(r)
	var userID *int64
	if ok {
		userID = &principal.UserID
	}
	return h.recordAudit(w, r, audit.Event{
		UserID:     userID,
		Action:     audit.ActionProcessList,
		Target:     "host.processes",
		Outcome:    audit.OutcomeSuccess,
		ReasonCode: "process_list",
	})
}
