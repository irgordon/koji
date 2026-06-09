package http

import (
	"net/http"

	"koji/internal/audit"
	"koji/internal/caps"
	"koji/internal/jobs"
)

func (h protectedHandlers) recordCapabilityDenied(w http.ResponseWriter, r *http.Request, userID *int64, capability caps.Capability, target string) bool {
	return h.recordAudit(w, r, audit.Event{
		UserID:     userID,
		Action:     audit.ActionCapabilityDeny,
		Target:     target,
		Outcome:    audit.OutcomeDenied,
		ReasonCode: string(capability),
	})
}

func (h protectedHandlers) recordDevBypass(w http.ResponseWriter, r *http.Request, capability caps.Capability, target string) bool {
	return h.recordAudit(w, r, audit.Event{
		Action:     audit.ActionCapabilityPass,
		Target:     target,
		Outcome:    audit.OutcomeAccepted,
		ReasonCode: string(capability),
		DevBypass:  true,
	})
}

func (h protectedHandlers) recordServiceControl(w http.ResponseWriter, r *http.Request, target string, outcome string, reason string) bool {
	principal, ok := principalFromRequest(r)
	var userID *int64
	if ok {
		userID = &principal.UserID
	}
	return h.recordAudit(w, r, audit.Event{
		UserID:     userID,
		Action:     audit.ActionServiceControl,
		Target:     target,
		Outcome:    outcome,
		ReasonCode: reason,
	})
}

func (h protectedHandlers) recordJobCreated(w http.ResponseWriter, r *http.Request, userID *int64, job jobs.Job, target string) bool {
	return h.recordAudit(w, r, audit.Event{
		UserID:     userID,
		Action:     audit.ActionJobCreated,
		Target:     target,
		Outcome:    audit.OutcomeAccepted,
		ReasonCode: job.Status,
	})
}

func (h protectedHandlers) recordJobStatusChanged(w http.ResponseWriter, r *http.Request, userID *int64, job jobs.Job, target string) bool {
	return h.recordAudit(w, r, audit.Event{
		UserID:     userID,
		Action:     audit.ActionJobStatus,
		Target:     target,
		Outcome:    audit.OutcomeAccepted,
		ReasonCode: job.Status,
	})
}

func (h protectedHandlers) recordJobViewed(w http.ResponseWriter, r *http.Request, target string) bool {
	principal, ok := principalFromRequest(r)
	var userID *int64
	if ok {
		userID = &principal.UserID
	}
	return h.recordAudit(w, r, audit.Event{
		UserID:     userID,
		Action:     audit.ActionJobViewed,
		Target:     target,
		Outcome:    audit.OutcomeSuccess,
		ReasonCode: "jobs_read",
	})
}

func (h protectedHandlers) recordJobDecision(w http.ResponseWriter, r *http.Request, userID *int64, action string, outcome string, job jobs.Job) bool {
	return h.recordAudit(w, r, audit.Event{
		UserID:     userID,
		Action:     action,
		Target:     "jobs:" + job.ID,
		Outcome:    outcome,
		ReasonCode: job.Status,
	})
}

func (h protectedHandlers) recordJobApprovalDenied(w http.ResponseWriter, r *http.Request, userID *int64, target string, reason string) bool {
	return h.recordAudit(w, r, audit.Event{
		UserID:     userID,
		Action:     audit.ActionJobApproveDeny,
		Target:     target,
		Outcome:    audit.OutcomeDenied,
		ReasonCode: reason,
	})
}

func (h protectedHandlers) recordAudit(w http.ResponseWriter, r *http.Request, event audit.Event) bool {
	event.RequestID = requestID(r)
	event.RemoteAddr = r.RemoteAddr
	if err := h.audit.Record(r.Context(), event); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Audit recording failed")
		return false
	}
	return true
}
