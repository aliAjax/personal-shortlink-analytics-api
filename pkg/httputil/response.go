package httputil

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

// WriteJSON serializes payload as JSON into the response. A per-call buffer is
// used instead of a shared package-level one so concurrent handlers never cross
// contaminate each other's bodies.
func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	var responseBuffer bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&responseBuffer).Encode(payload); err != nil {
			http.Error(w, `{"error":"failed to encode response"}`, http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(status)
	if _, err := w.Write(responseBuffer.Bytes()); err != nil {
		// The header is already written, so log-only is the only safe fallback.
		http.Error(w, `{"error":"failed to encode response"}`, http.StatusInternalServerError)
	}
}

func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, ErrorResponse{Error: message})
}

func DecodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(target); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return false
	}
	return true
}
