package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"koji/internal/audit"
	"koji/internal/auth"
)

type credentialRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func handleBootstrap(store *auth.Store, auditStore *audit.Store, devMode bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request, ok := decodeCredentialRequest(w, r)
		if !ok {
			recordAuthAudit(auditStore, r, nil, audit.ActionBootstrap, audit.OutcomeFailure, "invalid_request", devMode)
			return
		}

		session, err := store.Bootstrap(r.Context(), request.Username, request.Password)
		if err != nil {
			recordAuthAudit(auditStore, r, nil, audit.ActionBootstrap, audit.OutcomeFailure, bootstrapReason(err), devMode)
			writeBootstrapError(w, err)
			return
		}

		recordAuthAudit(auditStore, r, &session.UserID, audit.ActionBootstrap, audit.OutcomeSuccess, "bootstrap_created", devMode)
		writeAuthenticatedSession(w, session, devMode)
	}
}

func handleLogin(store *auth.Store, auditStore *audit.Store, devMode bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request, ok := decodeCredentialRequest(w, r)
		if !ok {
			recordAuthAudit(auditStore, r, nil, audit.ActionLogin, audit.OutcomeFailure, "invalid_request", devMode)
			return
		}

		session, err := store.Login(r.Context(), request.Username, request.Password)
		if err != nil {
			recordAuthAudit(auditStore, r, nil, audit.ActionLogin, audit.OutcomeFailure, "invalid_credentials", devMode)
			writeJSONError(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}

		recordAuthAudit(auditStore, r, &session.UserID, audit.ActionLogin, audit.OutcomeSuccess, "session_created", devMode)
		writeAuthenticatedSession(w, session, devMode)
	}
}

func handleLogout(store *auth.Store, auditStore *audit.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionID, ok := sessionIDFromRequest(r)
		if !ok {
			recordAuthAudit(auditStore, r, nil, audit.ActionLogout, audit.OutcomeFailure, "missing_session", false)
			writeJSONError(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		if err := store.ValidateCSRF(r.Context(), sessionID, r.Header.Get(auth.CSRFHeaderName)); err != nil {
			recordAuthAudit(auditStore, r, nil, audit.ActionLogout, audit.OutcomeFailure, "csrf_denied", false)
			writeJSONError(w, authStatusCode(err), "CSRF token required")
			return
		}

		principal, err := store.ValidateSession(r.Context(), sessionID)
		if err != nil {
			recordAuthAudit(auditStore, r, nil, audit.ActionLogout, audit.OutcomeFailure, "invalid_session", false)
			writeJSONError(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		if err := store.RevokeSession(r.Context(), sessionID); err != nil {
			recordAuthAudit(auditStore, r, &principal.UserID, audit.ActionLogout, audit.OutcomeFailure, "revoke_failed", false)
			writeJSONError(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		recordAuthAudit(auditStore, r, &principal.UserID, audit.ActionLogout, audit.OutcomeSuccess, "session_revoked", false)
		clearAuthCookies(w)
		writeJSONStatus(w, http.StatusOK, "logged_out")
	}
}

func handleSessionStatus(store *auth.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if sessionID, ok := sessionIDFromRequest(r); ok {
			if principal, err := store.ValidateSession(r.Context(), sessionID); err == nil {
				writeJSON(w, http.StatusOK, map[string]any{
					"authenticated":     true,
					"username":          principal.Username,
					"bootstrapRequired": false,
				})
				return
			}
		}

		bootstrapRequired := bootstrapRequired(r, store)
		writeJSON(w, http.StatusOK, map[string]any{
			"authenticated":     false,
			"bootstrapRequired": bootstrapRequired,
		})
	}
}

func decodeCredentialRequest(w http.ResponseWriter, r *http.Request) (credentialRequest, bool) {
	var request credentialRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return credentialRequest{}, false
	}
	return request, true
}

func writeBootstrapError(w http.ResponseWriter, err error) {
	if errors.Is(err, auth.ErrBootstrapDisabled) {
		writeJSONError(w, http.StatusConflict, "Bootstrap disabled")
		return
	}
	writeJSONError(w, http.StatusBadRequest, "Bootstrap failed")
}

func bootstrapReason(err error) string {
	if errors.Is(err, auth.ErrBootstrapDisabled) {
		return "bootstrap_disabled"
	}
	if errors.Is(err, auth.ErrInvalidCredential) {
		return "invalid_credentials"
	}
	return "bootstrap_failed"
}

func recordAuthAudit(store *audit.Store, r *http.Request, userID *int64, action string, outcome string, reason string, devBypass bool) {
	if store == nil {
		return
	}
	_ = store.Record(r.Context(), audit.Event{
		UserID:     userID,
		Action:     action,
		Target:     "auth",
		Outcome:    outcome,
		ReasonCode: reason,
		RequestID:  requestID(r),
		RemoteAddr: r.RemoteAddr,
		DevBypass:  devBypass,
	})
}

func writeAuthenticatedSession(w http.ResponseWriter, session auth.Session, devMode bool) {
	setAuthCookies(w, session, devMode)
	writeJSON(w, http.StatusOK, map[string]any{
		"authenticated": true,
		"username":      session.Username,
		"csrfToken":     session.CSRFToken,
	})
}

func setAuthCookies(w http.ResponseWriter, session auth.Session, devMode bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     auth.SessionCookieName,
		Value:    session.ID,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   !devMode,
		SameSite: http.SameSiteStrictMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     auth.CSRFCookieName,
		Value:    session.CSRFToken,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: false,
		Secure:   !devMode,
		SameSite: http.SameSiteStrictMode,
	})
}

func clearAuthCookies(w http.ResponseWriter) {
	expired := time.Unix(0, 0)
	http.SetCookie(w, &http.Cookie{Name: auth.SessionCookieName, Path: "/", Expires: expired, MaxAge: -1})
	http.SetCookie(w, &http.Cookie{Name: auth.CSRFCookieName, Path: "/", Expires: expired, MaxAge: -1})
}

func bootstrapRequired(r *http.Request, store *auth.Store) bool {
	hasUsers, err := store.HasUsers(r.Context())
	return err == nil && !hasUsers
}
