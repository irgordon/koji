package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"koji/internal/audit"
	"koji/internal/caps"
	"koji/internal/identity"
)

type adminUserCreateRequest struct {
	Username string `json:"username"`
}

type capabilityRequest struct {
	Capability string `json:"capability"`
}

type adminUsersResponse struct {
	Users []identity.User `json:"users"`
}

type capabilitiesResponse struct {
	Capabilities []string `json:"capabilities"`
}

func (h protectedHandlers) handleAdminUsersList(w http.ResponseWriter, r *http.Request) {
	if !h.requireCapability(w, r, caps.IdentityUsersManage, "identity.users") {
		return
	}
	users, err := h.identity.ListUsers(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "List users failed")
		return
	}
	writeJSONValue(w, http.StatusOK, adminUsersResponse{Users: users})
}

func (h protectedHandlers) handleAdminUserCreate(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requireIdentityAdmin(w, r)
	if !ok {
		return
	}
	var request adminUserCreateRequest
	if !decodeAdminJSON(w, r, &request) {
		return
	}
	user, err := h.identity.CreateManagedUser(r.Context(), request.Username)
	if err != nil {
		writeAdminError(w, err)
		return
	}
	h.recordIdentityAudit(r, principal.UserID, audit.ActionIdentityUserCreated, userTarget(user.ID), "user_created")
	writeJSONValue(w, http.StatusCreated, user)
}

func (h protectedHandlers) handleAdminUserDisable(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requireIdentityAdmin(w, r)
	if !ok {
		return
	}
	id, ok := userIDPathValue(w, r)
	if !ok {
		return
	}
	user, err := h.identity.DisableUser(r.Context(), id)
	if err != nil {
		h.writeAdminDecisionError(w, r, principal.UserID, userTarget(id), err)
		return
	}
	h.recordIdentityAudit(r, principal.UserID, audit.ActionIdentityUserDisabled, userTarget(user.ID), "user_disabled")
	writeJSONValue(w, http.StatusOK, user)
}

func (h protectedHandlers) handleAdminUserEnable(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requireIdentityAdmin(w, r)
	if !ok {
		return
	}
	id, ok := userIDPathValue(w, r)
	if !ok {
		return
	}
	user, err := h.identity.EnableUser(r.Context(), id)
	if err != nil {
		writeAdminError(w, err)
		return
	}
	h.recordIdentityAudit(r, principal.UserID, audit.ActionIdentityUserEnabled, userTarget(user.ID), "user_enabled")
	writeJSONValue(w, http.StatusOK, user)
}

func (h protectedHandlers) handleAdminUserCapabilities(w http.ResponseWriter, r *http.Request) {
	if !h.requireCapability(w, r, caps.IdentityUsersManage, "identity.users") {
		return
	}
	id, ok := userIDPathValue(w, r)
	if !ok {
		return
	}
	assigned, err := h.identity.ListUserCapabilities(r.Context(), id)
	if err != nil {
		writeAdminError(w, err)
		return
	}
	available, err := h.identity.ListAvailableCapabilities(r.Context())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "List capabilities failed")
		return
	}
	writeJSONValue(w, http.StatusOK, map[string]any{
		"capabilities": assigned,
		"available":    available,
	})
}

func (h protectedHandlers) handleAdminCapabilityGrant(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requireIdentityAdmin(w, r)
	if !ok {
		return
	}
	id, ok := userIDPathValue(w, r)
	if !ok {
		return
	}
	var request capabilityRequest
	if !decodeAdminJSON(w, r, &request) {
		return
	}
	capabilities, err := h.identity.GrantCapability(r.Context(), id, request.Capability)
	if err != nil {
		writeAdminError(w, err)
		return
	}
	h.recordIdentityAudit(r, principal.UserID, audit.ActionIdentityCapabilityGranted, userTarget(id), request.Capability)
	writeJSONValue(w, http.StatusOK, capabilitiesResponse{Capabilities: capabilities})
}

func (h protectedHandlers) handleAdminCapabilityRevoke(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requireIdentityAdmin(w, r)
	if !ok {
		return
	}
	id, ok := userIDPathValue(w, r)
	if !ok {
		return
	}
	capability := r.PathValue("capability")
	capabilities, err := h.identity.RevokeCapability(r.Context(), id, capability)
	if err != nil {
		h.writeAdminDecisionError(w, r, principal.UserID, userTarget(id), err)
		return
	}
	h.recordIdentityAudit(r, principal.UserID, audit.ActionIdentityCapabilityRevoked, userTarget(id), capability)
	writeJSONValue(w, http.StatusOK, capabilitiesResponse{Capabilities: capabilities})
}

func (h protectedHandlers) handleAdminMagicTokenIssue(w http.ResponseWriter, r *http.Request) {
	principal, ok := h.requireIdentityAdmin(w, r)
	if !ok {
		return
	}
	id, ok := userIDPathValue(w, r)
	if !ok {
		return
	}
	token, err := h.identity.IssueMagicToken(r.Context(), id, principal.UserID, h.magicTokenTTL)
	if err != nil {
		writeAdminError(w, err)
		return
	}
	h.recordIdentityAudit(r, principal.UserID, audit.ActionIdentityMagicTokenIssued, userTarget(id), "token_issued")
	writeJSONValue(w, http.StatusCreated, token)
}

func (h protectedHandlers) requireIdentityAdmin(w http.ResponseWriter, r *http.Request) (principalForAdmin, bool) {
	if !h.requireCapability(w, r, caps.IdentityUsersManage, "identity.users") {
		return principalForAdmin{}, false
	}
	principal, ok := principalFromRequest(r)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "Authentication required")
		return principalForAdmin{}, false
	}
	return principalForAdmin{UserID: principal.UserID}, true
}

type principalForAdmin struct {
	UserID int64
}

func (h protectedHandlers) recordIdentityAudit(r *http.Request, userID int64, action string, target string, reason string) {
	_ = h.audit.Record(r.Context(), audit.Event{
		UserID:     &userID,
		Action:     action,
		Target:     target,
		Outcome:    audit.OutcomeSuccess,
		ReasonCode: reason,
		RequestID:  requestID(r),
		RemoteAddr: r.RemoteAddr,
	})
}

func (h protectedHandlers) writeAdminDecisionError(w http.ResponseWriter, r *http.Request, userID int64, target string, err error) {
	if errors.Is(err, identity.ErrSelfLockout) {
		_ = h.audit.Record(r.Context(), audit.Event{
			UserID:     &userID,
			Action:     audit.ActionIdentitySelfLockoutPrevent,
			Target:     target,
			Outcome:    audit.OutcomeDenied,
			ReasonCode: "self_lockout_prevented",
			RequestID:  requestID(r),
			RemoteAddr: r.RemoteAddr,
		})
		writeJSONError(w, http.StatusConflict, "self_lockout_prevented")
		return
	}
	writeAdminError(w, err)
}

func decodeAdminJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return false
	}
	return true
}

func userIDPathValue(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSONError(w, http.StatusBadRequest, "Invalid user id")
		return 0, false
	}
	return id, true
}

func userTarget(id int64) string {
	return "users:" + strconv.FormatInt(id, 10)
}

func writeAdminError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, identity.ErrUserNotFound):
		writeJSONError(w, http.StatusNotFound, "User not found")
	case errors.Is(err, identity.ErrUserDisabled):
		writeJSONError(w, http.StatusConflict, "User is disabled")
	case errors.Is(err, identity.ErrInvalidCapability):
		writeJSONError(w, http.StatusBadRequest, "Invalid capability")
	case errors.Is(err, identity.ErrSelfLockout):
		writeJSONError(w, http.StatusConflict, "self_lockout_prevented")
	default:
		writeJSONError(w, http.StatusBadRequest, "Identity request failed")
	}
}
