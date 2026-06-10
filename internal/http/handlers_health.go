package http

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"koji/internal/agent"
	"koji/internal/db"
	"koji/internal/observability"
)

const (
	healthStatusOK       = "ok"
	healthStatusDegraded = "degraded"
	healthStatusFail     = "fail"
	readinessTimeout     = 500 * time.Millisecond
)

type operationalHandlers struct {
	database        *sql.DB
	agentSocketPath string
	metrics         *observability.Registry
}

type healthResponse struct {
	Status string                 `json:"status"`
	Checks map[string]checkResult `json:"checks"`
}

type checkResult struct {
	Status string `json:"status"`
}

func (h operationalHandlers) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSONValue(w, http.StatusOK, healthResponse{
		Status: healthStatusOK,
		Checks: map[string]checkResult{
			"liveness": {Status: healthStatusOK},
		},
	})
}

func (h operationalHandlers) handleReady(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), readinessTimeout)
	defer cancel()

	h.metrics.Inc(observability.ReadinessChecksTotal)
	report := h.readinessReport(ctx)
	writeJSONValue(w, readinessStatusCode(report.Status), report)
}

func (h operationalHandlers) readinessReport(ctx context.Context) healthResponse {
	checks := map[string]checkResult{
		"db":         h.dbCheck(ctx),
		"migrations": h.migrationCheck(ctx),
		"agent":      h.agentCheck(ctx),
	}
	return healthResponse{
		Status: readinessStatus(checks),
		Checks: checks,
	}
}

func (h operationalHandlers) dbCheck(ctx context.Context) checkResult {
	if err := h.database.PingContext(ctx); err != nil {
		h.metrics.Inc(observability.ReadinessDBFailuresTotal)
		return checkResult{Status: healthStatusFail}
	}
	return checkResult{Status: healthStatusOK}
}

func (h operationalHandlers) migrationCheck(ctx context.Context) checkResult {
	if err := db.CheckMigrationsCurrent(ctx, h.database, db.InitialMigrations()); err != nil {
		h.metrics.Inc(observability.ReadinessMigrationFailuresTotal)
		return checkResult{Status: healthStatusFail}
	}
	return checkResult{Status: healthStatusOK}
}

func (h operationalHandlers) agentCheck(ctx context.Context) checkResult {
	if err := agent.CheckReachable(ctx, h.agentSocketPath); err != nil {
		h.metrics.Inc(observability.ReadinessAgentDegradedTotal)
		return checkResult{Status: healthStatusDegraded}
	}
	return checkResult{Status: healthStatusOK}
}

func readinessStatus(checks map[string]checkResult) string {
	if checkFailed(checks["db"]) || checkFailed(checks["migrations"]) {
		return healthStatusFail
	}
	if checkDegraded(checks["agent"]) {
		return healthStatusDegraded
	}
	return healthStatusOK
}

func readinessStatusCode(status string) int {
	if status == healthStatusFail {
		return http.StatusServiceUnavailable
	}
	return http.StatusOK
}

func checkFailed(check checkResult) bool {
	return check.Status == healthStatusFail
}

func checkDegraded(check checkResult) bool {
	return check.Status == healthStatusDegraded
}
