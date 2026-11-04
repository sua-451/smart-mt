// Package handlers contains all HTTP handler helpers shared across modules.
package handlers

import (
	"encoding/json"
	"net/http"
)

// SetJSONHeaders writes CORS and Content-Type headers for JSON API responses.
func SetJSONHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")
}

// HandleOptions sends a 200 for CORS pre-flight requests.
func HandleOptions(w http.ResponseWriter, r *http.Request) {
	SetJSONHeaders(w)
	w.WriteHeader(http.StatusOK)
}

// WriteJSON serialises v as JSON and writes it to w.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	SetJSONHeaders(w)
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// WriteError sends a JSON error envelope.
func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}

