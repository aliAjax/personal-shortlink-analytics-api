package httputil

import (
	"bytes"
	"encoding/json"
	"net/http"
)

var responseBuffer bytes.Buffer

type ErrorResponse struct {
	Error string `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	responseBuffer.Reset()
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
