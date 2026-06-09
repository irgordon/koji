package http

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"koji/internal/audit"
	"koji/internal/auth"
	"koji/internal/db"
)

func TestUnauthenticatedAPIRequestReturnsUnauthorized(t *testing.T) {
	store, cleanup := newTestAuthStore(t)
	defer cleanup()

	response := exerciseAuthGate(store, false, http.MethodGet, "/api/v1/metrics", nil, "")

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", response.Code)
	}
}

func TestUnauthenticatedProductionSPARequestReturnsUnauthorized(t *testing.T) {
	store, cleanup := newTestAuthStore(t)
	defer cleanup()

	response := exerciseAuthGate(store, false, http.MethodGet, "/", nil, "")

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", response.Code)
	}
}

func TestBootstrapOnlyWorksBeforeFirstUser(t *testing.T) {
	fixture := newTestFixture(t)
	store := fixture.authStore
	defer fixture.cleanup()

	first := exerciseBootstrap(store, "admin", "secret-password")
	second := exerciseBootstrap(store, "other", "secret-password")

	if first.Code != http.StatusOK {
		t.Fatalf("expected first bootstrap status 200, got %d", first.Code)
	}
	if second.Code != http.StatusConflict {
		t.Fatalf("expected second bootstrap status 409, got %d", second.Code)
	}
}

func TestLoginCreatesSession(t *testing.T) {
	fixture := newTestFixture(t)
	store := fixture.authStore
	defer fixture.cleanup()
	bootstrapSession(t, store)

	response := exerciseLogin(store, "admin", "secret-password")

	if response.Code != http.StatusOK {
		t.Fatalf("expected login status 200, got %d", response.Code)
	}
	if sessionCookie(response) == nil {
		t.Fatal("expected session cookie")
	}
	if !strings.Contains(response.Body.String(), "csrfToken") {
		t.Fatalf("expected csrf token response, got %q", response.Body.String())
	}
}

func TestLogoutInvalidatesSession(t *testing.T) {
	fixture := newTestFixture(t)
	store := fixture.authStore
	defer fixture.cleanup()
	session := bootstrapSession(t, store)

	response := exerciseLogout(store, session)

	if response.Code != http.StatusOK {
		t.Fatalf("expected logout status 200, got %d", response.Code)
	}
	if _, err := store.ValidateSession(context.Background(), session.ID); !errors.Is(err, auth.ErrInvalidSession) {
		t.Fatalf("expected invalid session after logout, got %v", err)
	}
}

func TestServiceControlIntentRequiresSessionAndCSRF(t *testing.T) {
	store, cleanup := newTestAuthStore(t)
	defer cleanup()
	session := bootstrapSession(t, store)

	unauthenticated := exerciseAuthGate(store, false, http.MethodPost, "/api/services/ssh.service/restart", nil, "")
	missingCSRF := exerciseAuthGate(store, false, http.MethodPost, "/api/services/ssh.service/restart", sessionCookieFor(session), "")
	valid := exerciseAuthGate(store, false, http.MethodPost, "/api/services/ssh.service/restart", sessionCookieFor(session), session.CSRFToken)

	if unauthenticated.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthenticated status 401, got %d", unauthenticated.Code)
	}
	if missingCSRF.Code != http.StatusForbidden {
		t.Fatalf("expected missing csrf status 403, got %d", missingCSRF.Code)
	}
	if valid.Code != http.StatusOK {
		t.Fatalf("expected valid status 200, got %d", valid.Code)
	}
}

func TestDevModeBypassIsExplicit(t *testing.T) {
	store, cleanup := newTestAuthStore(t)
	defer cleanup()

	response := exerciseAuthGate(store, true, http.MethodGet, "/api/v1/metrics", nil, "")

	if response.Code != http.StatusOK {
		t.Fatalf("expected dev bypass status 200, got %d", response.Code)
	}
	if response.Header().Get("X-Koji-Auth-Bypass") != "dev" {
		t.Fatalf("expected explicit dev bypass header")
	}
}

func TestProductionSessionCookieIsHardened(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	bootstrapSession(t, fixture.authStore)

	response := exerciseLoginWithMode(fixture.authStore, "admin", "secret-password", false)
	cookie := sessionCookie(response)
	if cookie == nil {
		t.Fatal("expected session cookie")
	}
	if !cookie.Secure {
		t.Fatal("expected production session cookie to be secure")
	}
	if !cookie.HttpOnly {
		t.Fatal("expected production session cookie to be http only")
	}
	if cookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("expected strict same-site cookie, got %v", cookie.SameSite)
	}
	if cookie.Path != "/" {
		t.Fatalf("expected root cookie path, got %q", cookie.Path)
	}
}

func TestDevSessionCookieIsExplicitlyInsecureForLocalHTTP(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	bootstrapSession(t, fixture.authStore)

	response := exerciseLoginWithMode(fixture.authStore, "admin", "secret-password", true)
	cookie := sessionCookie(response)
	if cookie == nil {
		t.Fatal("expected session cookie")
	}
	if cookie.Secure {
		t.Fatal("expected dev session cookie to allow local http")
	}
	if !cookie.HttpOnly {
		t.Fatal("expected dev session cookie to remain http only")
	}
	if cookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("expected strict same-site cookie, got %v", cookie.SameSite)
	}
}

func newTestAuthStore(t *testing.T) (*auth.Store, func()) {
	fixture := newTestFixture(t)
	return fixture.authStore, fixture.cleanup
}

type testFixture struct {
	database   *sql.DB
	authStore  *auth.Store
	auditStore *audit.Store
	cleanup    func()
}

func newTestFixture(t *testing.T) testFixture {
	t.Helper()

	database, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "koji.db"))
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	return testFixture{
		database:   database,
		authStore:  auth.NewStore(database),
		auditStore: audit.NewStore(database),
		cleanup: func() {
			_ = database.Close()
		},
	}
}

func exerciseAuthGate(store *auth.Store, devMode bool, method string, path string, cookie *http.Cookie, csrfToken string) *httptest.ResponseRecorder {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := AuthGateMiddleware(store, devMode)(next)
	request := httptest.NewRequest(method, path, nil)
	if cookie != nil {
		request.AddCookie(cookie)
	}
	if csrfToken != "" {
		request.Header.Set(auth.CSRFHeaderName, csrfToken)
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func exerciseBootstrap(store *auth.Store, username string, password string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/api/bootstrap", credentialBody(username, password))
	response := httptest.NewRecorder()
	handleBootstrap(store, nil, true).ServeHTTP(response, request)
	return response
}

func exerciseLogin(store *auth.Store, username string, password string) *httptest.ResponseRecorder {
	return exerciseLoginWithMode(store, username, password, true)
}

func exerciseLoginWithMode(store *auth.Store, username string, password string, devMode bool) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/api/login", credentialBody(username, password))
	response := httptest.NewRecorder()
	handleLogin(store, nil, devMode).ServeHTTP(response, request)
	return response
}

func exerciseLogout(store *auth.Store, session auth.Session) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/api/logout", nil)
	request.AddCookie(sessionCookieFor(session))
	request.Header.Set(auth.CSRFHeaderName, session.CSRFToken)

	response := httptest.NewRecorder()
	handleLogout(store, nil).ServeHTTP(response, request)
	return response
}

func bootstrapSession(t *testing.T, store *auth.Store) auth.Session {
	t.Helper()

	session, err := store.Bootstrap(context.Background(), "admin", "secret-password")
	if err != nil {
		t.Fatalf("bootstrap session: %v", err)
	}
	return session
}

func credentialBody(username string, password string) *bytes.Reader {
	return bytes.NewReader([]byte(`{"username":"` + username + `","password":"` + password + `"}`))
}

func sessionCookieFor(session auth.Session) *http.Cookie {
	return &http.Cookie{Name: auth.SessionCookieName, Value: session.ID}
}

func sessionCookie(response *httptest.ResponseRecorder) *http.Cookie {
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == auth.SessionCookieName {
			return cookie
		}
	}
	return nil
}
