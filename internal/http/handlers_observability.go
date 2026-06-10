package http

import (
	"net/http"

	"koji/internal/caps"
)

func (h protectedHandlers) handleObservabilityMetrics(w http.ResponseWriter, r *http.Request) {
	if !h.requireCapability(w, r, caps.ObservabilityRead, "observability.metrics") {
		return
	}

	snapshot, err := h.metrics.Snapshot(r.Context(), h.database)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to read observability metrics")
		return
	}
	writeJSONValue(w, http.StatusOK, snapshot)
}
