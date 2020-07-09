package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

type jsonError struct {
	Error       string `json:"error"`
	Description string `json:"error_description"`
}

func writeJSONError(w http.ResponseWriter, error string, description string, statusCode int) {
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(jsonError{
		Error:       error,
		Description: description,
	}); err != nil {
		log.Println("handler/middleware failed to write response:", err)
	}
}
