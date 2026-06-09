package http

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"koji/internal/config"
)

func TestAuthenticatedSPAServesFromConfiguredAbsolutePath(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	session := bootstrapSession(t, fixture.authStore)
	staticDir := writeStaticFixture(t)

	handler := newProductionTestRouter(t, fixture, staticDir)
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.AddCookie(sessionCookieFor(session))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	if response.Body.String() != "<!doctype html><title>Koji</title>" {
		t.Fatalf("unexpected body %q", response.Body.String())
	}
	assertProductionSecurityHeaders(t, response)
}

func TestUnauthenticatedSPADenied(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	staticDir := writeStaticFixture(t)

	handler := newProductionTestRouter(t, fixture, staticDir)
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", response.Code)
	}
	assertProductionSecurityHeaders(t, response)
}

func TestAPIResponsesIncludeSecurityAndNoStoreHeaders(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	staticDir := writeStaticFixture(t)

	handler := newProductionTestRouter(t, fixture, staticDir)
	request := httptest.NewRequest(http.MethodGet, "/api/session", nil)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.Code)
	}
	assertProductionSecurityHeaders(t, response)
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("expected no-store cache header, got %q", response.Header().Get("Cache-Control"))
	}
}

func TestPathTraversalAttemptFailsSafely(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	session := bootstrapSession(t, fixture.authStore)
	staticDir := writeStaticFixture(t)

	handler := newProductionTestRouter(t, fixture, staticDir)
	request := httptest.NewRequest(http.MethodGet, "/../secret", nil)
	request.AddCookie(sessionCookieFor(session))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code == http.StatusOK {
		t.Fatalf("expected traversal attempt to fail safely, got %d", response.Code)
	}
}

func newProductionTestRouter(t *testing.T, fixture testFixture, staticAssetDir string) http.Handler {
	t.Helper()

	cfg := config.NewDefaultConfig()
	cfg.StaticAssetDir = staticAssetDir

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/session", handleSessionStatus(fixture.authStore))
	if err := registerStaticRoutes(mux, cfg); err != nil {
		t.Fatalf("register static routes: %v", err)
	}
	return applyMiddlewareChain(mux, fixture.authStore, cfg.DevMode)
}

func writeStaticFixture(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "assets"), 0755); err != nil {
		t.Fatalf("create assets dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<!doctype html><title>Koji</title>"), 0644); err != nil {
		t.Fatalf("write index: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "assets", "app.js"), []byte("console.log('koji')"), 0644); err != nil {
		t.Fatalf("write asset: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "secret"), []byte("secret"), 0644); err != nil {
		t.Fatalf("write secret: %v", err)
	}
	return dir
}

func assertProductionSecurityHeaders(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()

	expected := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"Referrer-Policy":        "no-referrer",
		"X-Frame-Options":        "DENY",
	}
	for header, value := range expected {
		if response.Header().Get(header) != value {
			t.Fatalf("expected %s %q, got %q", header, value, response.Header().Get(header))
		}
	}
	if response.Header().Get("Content-Security-Policy") == "" {
		t.Fatal("expected content security policy header")
	}
	if response.Header().Get("Permissions-Policy") == "" {
		t.Fatal("expected permissions policy header")
	}
}
