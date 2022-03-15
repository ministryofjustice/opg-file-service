package internal

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/ministryofjustice/opg-go-common/logging"
)

type jsonError struct {
	Error       string `json:"error"`
	Description string `json:"error_description"`
}

func WriteJSONError(w http.ResponseWriter, error string, description string, statusCode int) {
	l := logging.New(os.Stdout, "opg-file-service")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(jsonError{
		Error:       error,
		Description: description,
	}); err != nil {
		l.Print("handler/middleware failed to write response:", err)
	}
}
