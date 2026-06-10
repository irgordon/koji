package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"koji/internal/audit"
	"koji/internal/auth"
	"koji/internal/caps"
	"koji/internal/identity"
)

func TestBootstrapCreatesSuperAdminWithIdentityCapability(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	session := bootstrapSession(t, fixture.authStore)

	allowed, err := caps.NewStore(fixture.database).UserHasCapability(context.Background(), session.UserID, caps.IdentityUsersManage)
	if err != nil {
		t.Fatalf("lookup capability: %v", err)
	}
	if !allowed {
		t.Fatal("expected bootstrap user to receive identity.users.manage")
	}
}

func TestAdminUsersRequiresIdentityCapability(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	session := bootstrapSession(t, fixture.authStore)
	revokeCapabilityForTest(t, fixture, session.UserID, caps.IdentityUsersManage)

	response := exerciseAdminUsersList(fixture, session)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", response.Code)
	}
}

func TestAdminCanCreateManagedUserAndIssueMagicToken(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	admin := bootstrapSession(t, fixture.authStore)

	createResponse := exerciseAdminCreateUser(fixture, admin, "alice")
	user := decodeIdentityUser(t, createResponse)
	tokenResponse := exerciseAdminIssueMagicToken(fixture, admin, user.ID)
	token := decodeMagicToken(t, tokenResponse)

	if createResponse.Code != http.StatusCreated {
		t.Fatalf("expected user create 201, got %d", createResponse.Code)
	}
	if user.Username != "alice" || user.IsSuperAdmin || user.Disabled {
		t.Fatalf("unexpected managed user: %#v", user)
	}
	if tokenResponse.Code != http.StatusCreated {
		t.Fatalf("expected token create 201, got %d", tokenResponse.Code)
	}
	if token.Token == "" || token.ExpiresAt == "" {
		t.Fatalf("expected one-time token and expiry, got %#v", token)
	}
	if strings.Contains(tokenResponse.Body.String(), "token_hash") {
		t.Fatalf("raw response must not expose token hash: %s", tokenResponse.Body.String())
	}
}

func TestManagedUserPasswordLoginIsDenied(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	admin := bootstrapSession(t, fixture.authStore)
	user := decodeIdentityUser(t, exerciseAdminCreateUser(fixture, admin, "alice"))

	response := exerciseLogin(fixture.authStore, user.Username, "anything")

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected password login denied, got %d", response.Code)
	}
}

func TestMagicTokenLoginConsumesTokenOnce(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	admin := bootstrapSession(t, fixture.authStore)
	user := decodeIdentityUser(t, exerciseAdminCreateUser(fixture, admin, "alice"))
	token := decodeMagicToken(t, exerciseAdminIssueMagicToken(fixture, admin, user.ID))

	first := exerciseMagicTokenLogin(fixture.authStore, token.Token)
	second := exerciseMagicTokenLogin(fixture.authStore, token.Token)

	if first.Code != http.StatusOK {
		t.Fatalf("expected first magic token login 200, got %d", first.Code)
	}
	if second.Code != http.StatusUnauthorized {
		t.Fatalf("expected consumed token denied, got %d", second.Code)
	}
}

func TestCannotIssueMagicTokenForDisabledUser(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	admin := bootstrapSession(t, fixture.authStore)
	user := decodeIdentityUser(t, exerciseAdminCreateUser(fixture, admin, "alice"))

	exerciseAdminDisableUser(fixture, admin, user.ID)
	response := exerciseAdminIssueMagicToken(fixture, admin, user.ID)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected disabled user token issue conflict, got %d", response.Code)
	}
}

func TestCannotDisableFinalIdentityManager(t *testing.T) {
	fixture := newTestFixture(t)
	defer fixture.cleanup()
	admin := bootstrapSession(t, fixture.authStore)

	response := exerciseAdminDisableUser(fixture, admin, admin.UserID)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected self-lockout conflict, got %d", response.Code)
	}
	if !auditEventExists(t, fixture.database, audit.ActionIdentitySelfLockoutPrevent, audit.OutcomeDenied, "self_lockout_prevented") {
		t.Fatal("expected self-lockout audit event")
	}
}

func exerciseAdminUsersList(fixture testFixture, session auth.Session) *httptest.ResponseRecorder {
	request := authenticatedJSONRequest(http.MethodGet, "/api/admin/users", session, "")
	response := httptest.NewRecorder()
	fixtureAdminHandlers(fixture).handleAdminUsersList(response, request)
	return response
}

func exerciseAdminCreateUser(fixture testFixture, session auth.Session, username string) *httptest.ResponseRecorder {
	request := authenticatedJSONRequest(http.MethodPost, "/api/admin/users", session, `{"username":"`+username+`"}`)
	response := httptest.NewRecorder()
	fixtureAdminHandlers(fixture).handleAdminUserCreate(response, request)
	return response
}

func exerciseAdminIssueMagicToken(fixture testFixture, session auth.Session, id int64) *httptest.ResponseRecorder {
	request := authenticatedJSONRequest(http.MethodPost, "/api/admin/users/{id}/magic-token", session, "")
	request.SetPathValue("id", intString(id))
	response := httptest.NewRecorder()
	fixtureAdminHandlers(fixture).handleAdminMagicTokenIssue(response, request)
	return response
}

func exerciseAdminDisableUser(fixture testFixture, session auth.Session, id int64) *httptest.ResponseRecorder {
	request := authenticatedJSONRequest(http.MethodPost, "/api/admin/users/{id}/disable", session, "")
	request.SetPathValue("id", intString(id))
	response := httptest.NewRecorder()
	fixtureAdminHandlers(fixture).handleAdminUserDisable(response, request)
	return response
}

func exerciseMagicTokenLogin(store *auth.Store, token string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/api/login/magic-token", strings.NewReader(`{"token":"`+token+`"}`))
	response := httptest.NewRecorder()
	handleMagicTokenLogin(store, nil, true).ServeHTTP(response, request)
	return response
}

func fixtureAdminHandlers(fixture testFixture) protectedHandlers {
	return protectedHandlers{
		caps:          caps.NewStore(fixture.database),
		audit:         audit.NewStore(fixture.database),
		identity:      identity.NewStore(fixture.database),
		magicTokenTTL: auth.DefaultSessionPolicy().IdleTimeout / 2,
	}
}

func decodeIdentityUser(t *testing.T, response *httptest.ResponseRecorder) identity.User {
	t.Helper()

	var user identity.User
	if err := json.NewDecoder(response.Body).Decode(&user); err != nil {
		t.Fatalf("decode user: %v", err)
	}
	return user
}

func decodeMagicToken(t *testing.T, response *httptest.ResponseRecorder) identity.MagicToken {
	t.Helper()

	var token identity.MagicToken
	if err := json.NewDecoder(response.Body).Decode(&token); err != nil {
		t.Fatalf("decode magic token: %v", err)
	}
	return token
}

func revokeCapabilityForTest(t *testing.T, fixture testFixture, userID int64, capability caps.Capability) {
	t.Helper()

	_, err := fixture.database.Exec("DELETE FROM user_capabilities WHERE user_id = ? AND capability_name = ?", userID, string(capability))
	if err != nil {
		t.Fatalf("revoke test capability: %v", err)
	}
}

func intString(value int64) string {
	return strconv.FormatInt(value, 10)
}
