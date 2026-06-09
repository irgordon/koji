package http

import (
	"errors"
	"net/http"

	"koji/internal/audit"
	"koji/internal/caps"
)

func (h protectedHandlers) requireCapability(w http.ResponseWriter, r *http.Request, capability caps.Capability, target string) bool {
	if h.devMode {
		return h.recordDevBypass(w, r, capability, target)
	}

	principal, ok := principalFromRequest(r)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "Authentication required")
		return false
	}
	err := h.caps.Require(r.Context(), principal.UserID, capability)
	if err == nil {
		return true
	}
	if errors.Is(err, caps.ErrCapabilityDenied) {
		h.recordCapabilityDenied(w, r, &principal.UserID, capability, target)
		writeJSONError(w, http.StatusForbidden, "Capability denied")
		return false
	}
	writeJSONError(w, http.StatusInternalServerError, "Capability check failed")
	return false
}

func (h protectedHandlers) requireServiceControlCapability(w http.ResponseWriter, r *http.Request, target string) bool {
	if h.devMode {
		return h.recordDevBypass(w, r, caps.HostServicesControl, target)
	}

	principal, ok := principalFromRequest(r)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "Authentication required")
		return false
	}
	err := h.caps.Require(r.Context(), principal.UserID, caps.HostServicesControl)
	if err == nil {
		return true
	}
	if !errors.Is(err, caps.ErrCapabilityDenied) {
		writeJSONError(w, http.StatusInternalServerError, "Capability check failed")
		return false
	}

	if !h.recordCapabilityDenied(w, r, &principal.UserID, caps.HostServicesControl, target) {
		return false
	}
	if !h.recordAudit(w, r, audit.Event{
		UserID:     &principal.UserID,
		Action:     audit.ActionServiceControl,
		Target:     target,
		Outcome:    audit.OutcomeDenied,
		ReasonCode: "capability_denied",
	}) {
		return false
	}
	writeJSONError(w, http.StatusForbidden, "Capability denied")
	return false
}
