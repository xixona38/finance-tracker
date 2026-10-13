package httpresponse

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// WriteJSON converts data to JSON and sends it with the requested HTTP status.
// If the data cannot be converted, it sends a plain-text 500 error instead.
// It logs encoding and write errors without trying to replace a response that has already started.
func WriteJSON(w http.ResponseWriter, statusCode int, data any, logger *slog.Logger) {
	body, err := json.Marshal(data)
	if err != nil {
		logger.Error("Failed to encode JSON response", "error", err)
		http.Error(w, "failed to prepare response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, err = w.Write(body)
	if err != nil {
		logger.Error("Failed to write HTTP response", "error", err)
	}
}

// WriteError sends the error message as JSON with the requested HTTP status.
// It passes the logger to WriteJSON to report failures while preparing or sending the response.
func WriteError(w http.ResponseWriter, statusCode int, msg string, logger *slog.Logger) {
	WriteJSON(w, statusCode, map[string]string{"error": msg}, logger)
}
