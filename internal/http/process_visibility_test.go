package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"koji/internal/audit"
	"koji/internal/caps"
	"koji/internal/config"
	"koji/internal/system"
)

func TestDefaultProcessPolicyRedactsSensitiveFields(t *testing.T) {
	views := summaryProcessPolicy().apply(testProcesses())

	if len(views) != 2 {
		t.Fatalf("expected two process views, got %d", len(views))
	}
	if views[0].UID != nil || views[0].PPID != nil || views[0].RSS != nil || views[0].MemoryPct != nil {
		t.Fatalf("expected sensitive fields to be redacted: %#v", views[0])
	}
}

func TestProcessPolicyOmitsCommandLineByDefault(t *testing.T) {
	views := processVisibilityPolicy{
		mode:               config.ProcessVisibilityAll,
		includeCommandLine: false,
		maxProcesses:       10,
	}.apply(testProcesses())

	if views[0].CommandLine != "" {
		t.Fatalf("expected command line to be omitted, got %q", views[0].CommandLine)
	}
}

func TestProcessPolicyIncludesCommandLineWhenExplicit(t *testing.T) {
	views := processVisibilityPolicy{
		mode:               config.ProcessVisibilityAll,
		includeCommandLine: true,
		maxProcesses:       10,
	}.apply(testProcesses())

	if views[0].CommandLine != "/usr/bin/sshd -D" {
		t.Fatalf("expected command line, got %q", views[0].CommandLine)
	}
}

func TestProcessPolicyEnforcesMaxProcesses(t *testing.T) {
	views := processVisibilityPolicy{
		mode:         config.ProcessVisibilitySummary,
		maxProcesses: 1,
	}.apply(testProcesses())

	if len(views) != 1 {
		t.Fatalf("expected one process, got %d", len(views))
	}
}

func TestProcessListingRequiresCapability(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	session := bootstrapSession(t, fixture.authStore)

	request := authenticatedRequest(http.MethodGet, "/api/v1/processes", session)
	response := httptest.NewRecorder()
	protectedHandler(fixture, false).handleProcessesListWithLister(testProcessLister)(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", response.Code)
	}
}

func testProcessLister() ([]system.ProcessInfo, error) {
	return testProcesses(), nil
}

func TestProcessListingAuditEventIsWritten(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	session := bootstrapSession(t, fixture.authStore)
	grantCapability(t, fixture.database, session.UserID, caps.HostProcessesRead)

	request := authenticatedRequest(http.MethodGet, "/api/v1/processes", session)
	response := httptest.NewRecorder()
	protectedHandler(fixture, false).handleProcessesListWithLister(testProcessLister)(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d with body %q", response.Code, response.Body.String())
	}
	if !processListAuditExists(t, fixture) {
		t.Fatal("expected process listing audit event")
	}
}

func summaryProcessPolicy() processVisibilityPolicy {
	return processVisibilityPolicy{
		mode:               config.ProcessVisibilitySummary,
		includeCommandLine: false,
		maxProcesses:       10,
	}
}

func testProcesses() []system.ProcessInfo {
	return []system.ProcessInfo{
		{
			PID:         100,
			Name:        "sshd",
			State:       "S",
			UID:         0,
			PPID:        1,
			CPUUser:     10,
			CPUSystem:   20,
			RSS:         4096,
			MemoryPct:   1.5,
			CommandLine: "/usr/bin/sshd -D",
		},
		{
			PID:         200,
			Name:        "kojid",
			State:       "R",
			UID:         1000,
			PPID:        1,
			CPUUser:     30,
			CPUSystem:   40,
			RSS:         8192,
			MemoryPct:   2.5,
			CommandLine: "/usr/bin/kojid",
		},
	}
}

func processListAuditExists(t *testing.T, fixture testFixture) bool {
	t.Helper()

	var count int
	err := fixture.database.QueryRowContext(context.Background(), `
SELECT COUNT(*)
FROM audit_events
WHERE action = ? AND outcome = ? AND target = ?`, audit.ActionProcessList, audit.OutcomeSuccess, "host.processes").Scan(&count)
	if err != nil {
		t.Fatalf("query process audit event: %v", err)
	}
	return count > 0
}
