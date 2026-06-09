package http

import (
	"net/http"

	"koji/internal/caps"
	"koji/internal/system"
)

func (h protectedHandlers) handleMetricsFetch(probe *system.Probe) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !h.requireCapability(w, r, caps.HostMetricsRead, "host.metrics") {
			return
		}

		metrics, err := probe.Collect()
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "Failed to collect system metrics")
			return
		}

		writeJSONValue(w, http.StatusOK, metrics)
	}
}

func (h protectedHandlers) handleDiskFetch(w http.ResponseWriter, r *http.Request) {
	if !h.requireCapability(w, r, caps.HostDiskRead, "host.disk") {
		return
	}

	metrics, err := system.CollectDiskMetrics("/")
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to read disk metrics")
		return
	}

	writeJSONValue(w, http.StatusOK, metrics)
}
