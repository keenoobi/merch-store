package utils

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type ErrorResponse struct {
	Errors string `json:"errors"`
}

func WriteError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(ErrorResponse{Errors: message}); err != nil {
		slog.Error("failed to encode JSON response")
	}

}
