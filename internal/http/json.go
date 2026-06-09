package http

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, code int, payload map[string]any) {
	writeJSONValue(w, code, payload)
}

func writeJSONValue(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeJSONStatus(w http.ResponseWriter, code int, status string) {
	writeJSONValue(w, code, map[string]string{"status": status})
}

func writeJSONError(w http.ResponseWriter, code int, message string) {
	writeJSONValue(w, code, map[string]string{"error": message})
}
