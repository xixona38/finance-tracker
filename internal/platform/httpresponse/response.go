package httpresponse

import (
	"encoding/json"
	"fmt"
	"net/http"
)

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

func WriteError(w http.ResponseWriter, statusCode int, msg string) {
	WriteJSON(w, statusCode, map[string]string{"error": msg})
}
