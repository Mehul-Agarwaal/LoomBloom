package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
)

type errorResponse struct {
	Error string `json:"error"`
}

func decodeJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, err error) {
	message := "internal server error"
	if status < http.StatusInternalServerError && err != nil {
		message = err.Error()
	}
	writeJSON(w, status, errorResponse{Error: message})
}

func required(value, field string) error {
	if value == "" {
		return errors.New(field + " is required")
	}
	return nil
}
