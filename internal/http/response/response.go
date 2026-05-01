package response

import (
	"encoding/json"
	"net/http"
)

type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ErrorBody struct {
	Error  string        `json:"error"`
	Errors []ErrorDetail `json:"errors,omitempty"`
}

func JSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(payload)
}

func Error(w http.ResponseWriter, statusCode int, message string, details []ErrorDetail) {
	JSON(w, statusCode, ErrorBody{
		Error:  message,
		Errors: details,
	})
}
