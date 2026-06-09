package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"koji/internal/audit"
	"koji/internal/caps"
	"koji/internal/jobs"
)

const maxJobDecisionReasonLength = 512

type jobsListResponse struct {
	Jobs []jobs.Job `json:"jobs"`
}

type jobDecisionRequest struct {
	Reason string `json:"reason"`
}

func (h protectedHandlers) handleJobsList(w http.ResponseWriter, r *http.Request) {
	if !h.requireCapability(w, r, caps.JobsRead, "jobs") {
		return
	}

	result, err := h.jobs.ListRecent(r.Context(), jobs.DefaultListLimit)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to list jobs")
		return
	}
	if !h.recordJobViewed(w, r, "jobs") {
		return
	}
	writeJSONValue(w, http.StatusOK, jobsListResponse{Jobs: result})
}

func (h protectedHandlers) handleJobDetail(w http.ResponseWriter, r *http.Request) {
	if !h.requireCapability(w, r, caps.JobsRead, "jobs") {
		return
	}

	job, err := h.jobs.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		writeJobReadError(w, err)
		return
	}
	if !h.recordJobViewed(w, r, "jobs:"+job.ID) {
		return
	}
	writeJSONValue(w, http.StatusOK, job)
}

func (h protectedHandlers) handleJobApprove(w http.ResponseWriter, r *http.Request) {
	h.handleJobDecision(w, r, approveJobDecision)
}

func (h protectedHandlers) handleJobReject(w http.ResponseWriter, r *http.Request) {
	h.handleJobDecision(w, r, rejectJobDecision)
}

func (h protectedHandlers) handleJobDecision(w http.ResponseWriter, r *http.Request, decision jobDecisionHandler) {
	principal, ok := h.requireJobApproval(w, r, r.PathValue("id"))
	if !ok {
		return
	}
	reason, ok := readJobDecisionReason(w, r)
	if !ok {
		return
	}
	job, err := decision.apply(r, jobs.DecisionRequest{
		JobID:     r.PathValue("id"),
		DecidedBy: principal.UserID,
		Reason:    reason,
	}, h.jobs)
	if err != nil {
		h.writeJobDecisionError(w, r, err, r.PathValue("id"), principal.UserID)
		return
	}
	if !h.recordJobDecision(w, r, &principal.UserID, decision.auditAction, decision.auditOutcome, job) {
		return
	}
	if !h.recordJobStatusChanged(w, r, &principal.UserID, job, "jobs:"+job.ID) {
		return
	}
	writeJSONValue(w, http.StatusOK, job)
}

func writeJobReadError(w http.ResponseWriter, err error) {
	if errors.Is(err, jobs.ErrJobNotFound) {
		writeJSONError(w, http.StatusNotFound, "Job not found")
		return
	}
	writeJSONError(w, http.StatusInternalServerError, "Failed to read job")
}

type jobDecisionHandler struct {
	auditAction  string
	auditOutcome string
	apply        func(*http.Request, jobs.DecisionRequest, *jobs.Store) (jobs.Job, error)
}

var approveJobDecision = jobDecisionHandler{
	auditAction:  audit.ActionJobApproved,
	auditOutcome: audit.OutcomeAccepted,
	apply: func(r *http.Request, request jobs.DecisionRequest, store *jobs.Store) (jobs.Job, error) {
		return store.Approve(r.Context(), request)
	},
}

var rejectJobDecision = jobDecisionHandler{
	auditAction:  audit.ActionJobRejected,
	auditOutcome: audit.OutcomeRejected,
	apply: func(r *http.Request, request jobs.DecisionRequest, store *jobs.Store) (jobs.Job, error) {
		return store.Reject(r.Context(), request)
	},
}

func readJobDecisionReason(w http.ResponseWriter, r *http.Request) (string, bool) {
	payload, ok := decodeJobDecisionRequest(w, r)
	if !ok {
		return "", false
	}
	reason := normalizeJobDecisionReason(payload.Reason)
	if len(reason) > maxJobDecisionReasonLength {
		writeJSONError(w, http.StatusBadRequest, "Decision reason is too long")
		return "", false
	}
	return reason, true
}

func decodeJobDecisionRequest(w http.ResponseWriter, r *http.Request) (jobDecisionRequest, bool) {
	if r.Body == nil {
		return jobDecisionRequest{}, true
	}
	var payload jobDecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid job decision request")
		return jobDecisionRequest{}, false
	}
	return payload, true
}

func normalizeJobDecisionReason(reason string) string {
	return strings.TrimSpace(reason)
}

func (h protectedHandlers) requireJobApproval(w http.ResponseWriter, r *http.Request, id string) (authPrincipal, bool) {
	principal, ok := principalFromRequest(r)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "Authentication required")
		return authPrincipal{}, false
	}
	if h.devMode {
		if !h.recordDevBypass(w, r, caps.JobsApprove, "jobs:"+id) {
			return authPrincipal{}, false
		}
		return authPrincipal{UserID: principal.UserID}, true
	}
	if err := h.caps.Require(r.Context(), principal.UserID, caps.JobsApprove); err != nil {
		return authPrincipal{}, h.handleJobApprovalCapabilityError(w, r, err, id, principal.UserID)
	}
	return authPrincipal{UserID: principal.UserID}, true
}

type authPrincipal struct {
	UserID int64
}

func (h protectedHandlers) handleJobApprovalCapabilityError(w http.ResponseWriter, r *http.Request, err error, id string, userID int64) bool {
	if !errors.Is(err, caps.ErrCapabilityDenied) {
		writeJSONError(w, http.StatusInternalServerError, "Capability check failed")
		return false
	}
	if !h.recordCapabilityDenied(w, r, &userID, caps.JobsApprove, "jobs:"+id) {
		return false
	}
	if !h.recordJobApprovalDenied(w, r, &userID, "jobs:"+id, "capability_denied") {
		return false
	}
	writeJSONError(w, http.StatusForbidden, "Capability denied")
	return false
}

func (h protectedHandlers) writeJobDecisionError(w http.ResponseWriter, r *http.Request, err error, id string, userID int64) {
	if errors.Is(err, jobs.ErrJobNotFound) {
		writeJSONError(w, http.StatusNotFound, "Job not found")
		return
	}
	if errors.Is(err, jobs.ErrInvalidJobTransition) {
		if !h.recordJobApprovalDenied(w, r, &userID, "jobs:"+id, "invalid_status") {
			return
		}
		writeJSONError(w, http.StatusConflict, "Job is not queued")
		return
	}
	writeJSONError(w, http.StatusInternalServerError, "Failed to decide job")
}
