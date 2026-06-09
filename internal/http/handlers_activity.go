package http

import (
	"net/http"

	"koji/internal/audit"
	"koji/internal/caps"
)

type activityResponse struct {
	Events []audit.ActivityEvent `json:"events"`
}

func (h protectedHandlers) handleActivityList(w http.ResponseWriter, r *http.Request) {
	if !h.requireCapability(w, r, caps.AuditEventsRead, "audit.events") {
		return
	}

	events, err := h.audit.ListRecent(r.Context(), audit.DefaultRecentLimit)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to list activity")
		return
	}

	writeJSONValue(w, http.StatusOK, activityResponse{Events: events})
}
