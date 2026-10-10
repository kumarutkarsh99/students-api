package response

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

type ErrorResponse struct {
	Error  string `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, status int, message string) error {
	return WriteJSON(w, status, ErrorResponse{Error: message})
}

func ValidationError(errs validator.ValidationErrors) ErrorResponse {
	var errMsgs []string
	for _, err := range errs {
		field := strings.ToLower(err.Field())
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, field+" is required")
		case "gt":
			errMsgs = append(errMsgs, field+" must be greater than zero")
		case "email":
			errMsgs = append(errMsgs, field+" must be a valid email address")
		default:
			errMsgs = append(errMsgs, field+" is invalid")
		}
	}
	return ErrorResponse{Error: strings.Join(errMsgs, ", ")}
}
