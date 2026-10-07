package httpresponse

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// WriteJSON converts data to JSON and sends it with the requested HTTP status.
// If the data cannot be converted, it sends a plain-text 500 error instead.
func WriteJSON(w http.ResponseWriter, statusCode int, data any) {
	body, err := json.Marshal(data)
	if err != nil {
		fmt.Print("тут нужно будет написать логгирование!")
		http.Error(w, "failed to prepare response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, err = w.Write(body)
	if err != nil {
		fmt.Println("здесь нужно логгирование")
	}
}

// WriteError sends the error message as JSON with the requested HTTP status.
func WriteError(w http.ResponseWriter, statusCode int, msg string) {
	WriteJSON(w, statusCode, map[string]string{"error": msg})
}
