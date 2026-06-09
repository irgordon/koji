package http

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"koji/internal/db"
)

func TestHealthzReturnsOKWithoutAuth(t *testing.T) {
	handler := publicOperationalTestHandler(operationalHandlers{})
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	assertHealthStatus(t, response, healthStatusOK)
}

func TestReadyzReportsFailWhenDBUnavailable(t *testing.T) {
	database := openHealthTestDB(t)
	_ = database.Close()
	handler := operationalHandlers{database: database, agentSocketPath: filepath.Join(t.TempDir(), "agent.sock")}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)

	handler.handleReady(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", response.Code)
	}
	assertHealthStatus(t, response, healthStatusFail)
}

func TestReadyzReportsDegradedWhenAgentUnavailable(t *testing.T) {
	database := openHealthTestDB(t)
	defer database.Close()
	handler := operationalHandlers{database: database, agentSocketPath: filepath.Join(t.TempDir(), "agent.sock")}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)

	handler.handleReady(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	assertHealthStatus(t, response, healthStatusDegraded)
}

func TestHealthEndpointsDoNotExposeProtectedTelemetry(t *testing.T) {
	database := openHealthTestDB(t)
	defer database.Close()
	handler := publicOperationalTestHandler(operationalHandlers{
		database:        database,
		agentSocketPath: filepath.Join(t.TempDir(), "agent.sock"),
	})

	healthBody := exerciseOperationalEndpoint(t, handler, "/healthz")
	readyBody := exerciseOperationalEndpoint(t, handler, "/readyz")

	assertNoProtectedTelemetry(t, healthBody)
	assertNoProtectedTelemetry(t, readyBody)
}

func publicOperationalTestHandler(operational operationalHandlers) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", operational.handleHealth)
	mux.HandleFunc("GET /readyz", operational.handleReady)
	return applyMiddlewareChain(mux, nil, false)
}

func openHealthTestDB(t *testing.T) *sql.DB {
	t.Helper()

	database, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "koji.db"))
	if err != nil {
		t.Fatalf("open health test db: %v", err)
	}
	return database
}

func exerciseOperationalEndpoint(t *testing.T, handler http.Handler, path string) string {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, path, nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code >= 500 && path == "/healthz" {
		t.Fatalf("unexpected health endpoint status %d", response.Code)
	}
	return response.Body.String()
}

func assertHealthStatus(t *testing.T, response *httptest.ResponseRecorder, status string) {
	t.Helper()

	var payload healthResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode health response: %v", err)
	}
	if payload.Status != status {
		t.Fatalf("expected status %q, got %q", status, payload.Status)
	}
}

func assertNoProtectedTelemetry(t *testing.T, body string) {
	t.Helper()

	protectedTokens := []string{
		"cpuUsage",
		"memTotal",
		"services",
		"processes",
		"sessions",
		"audit",
		"users",
	}
	for _, token := range protectedTokens {
		if strings.Contains(body, token) {
			t.Fatalf("health endpoint exposed protected token %q in %q", token, body)
		}
	}
}
