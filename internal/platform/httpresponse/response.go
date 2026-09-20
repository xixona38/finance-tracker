package httpresponse

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		fmt.Print("тут нужно будет написать логгирование!")
	}
}

func WriteError(w http.ResponseWriter, statusCode int, msg string) {
	WriteJSON(w, statusCode, map[string]string{"error": msg})
}
