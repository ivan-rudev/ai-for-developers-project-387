package middleware

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse — тело ошибки в формате ErrorResponse из OpenAPI.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// WriteError записывает JSON-ошибку ErrorResponse в response writer.
func WriteError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	//nolint:gosec // ошибка кодирования после WriteHeader невосстановима
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: code, Message: message})
}
