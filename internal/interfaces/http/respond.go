package http

import (
	"encoding/json"
	"errors"
	nethttp "net/http"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/interfaces/http/middleware"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/usecase"
)

// writeJSON записывает тело ответа в формате JSON.
func writeJSON(w nethttp.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	//nolint:gosec // ошибка кодирования после WriteHeader невосстановима
	_ = json.NewEncoder(w).Encode(v)
}

// writeError преобразует ошибку бизнес-логики в JSON ErrorResponse.
func writeError(w nethttp.ResponseWriter, err error) {
	status, code, message := statusOfError(err)
	middleware.WriteError(w, status, code, message)
}

// writeOwnerNotFound пишет 404 «owner not found», если ошибка вызвана отсутствием
// владельца, иначе передаёт ошибку дальше.
func writeOwnerNotFound(w nethttp.ResponseWriter, err error) {
	if errors.Is(err, domain.ErrNotFound) {
		middleware.WriteError(w, nethttp.StatusNotFound, "not_found", "owner not found")
		return
	}
	writeError(w, err)
}

// statusOfError сопоставляет доменные и use case ошибки с HTTP-статусом, кодом
// и человекочитаемым сообщением.
func statusOfError(err error) (int, string, string) {
	switch {
	case errors.Is(err, domain.ErrRateLimit):
		return nethttp.StatusTooManyRequests, "rate_limit", "too many requests"
	case errors.Is(err, domain.ErrNotFound):
		return nethttp.StatusNotFound, "not_found", messageOr(err, "not found")
	case errors.Is(err, domain.ErrOverlap):
		return nethttp.StatusConflict, "overlap", messageOr(err, "booking overlap")
	case errors.Is(err, domain.ErrSlotUnavailable):
		return nethttp.StatusConflict, "slot_unavailable", messageOr(err, "slot is unavailable")
	case errors.Is(err, domain.ErrConflict):
		return nethttp.StatusConflict, "conflict", messageOr(err, "conflict")
	case errors.Is(err, domain.ErrInvalidInput):
		return nethttp.StatusBadRequest, "invalid_input", messageOr(err, "invalid input")
	case errors.Is(err, domain.ErrRequestBodyTooLarge):
		return nethttp.StatusRequestEntityTooLarge, "request_body_too_large", "request body too large"
	default:
		return nethttp.StatusInternalServerError, "internal_error", "internal server error"
	}
}

// messageOr возвращает пользовательское сообщение ошибки или fallback.
func messageOr(err error, fallback string) string {
	if msg := usecase.Message(err); msg != "" {
		return msg
	}
	return fallback
}

// decodeJSON разбирает тело запроса как JSON.
func decodeJSON(r *nethttp.Request, v any) error {
	const maxBodySize = 1 << 20 // 1 MB
	if r.Body == nil {
		return domain.ErrInvalidInput
	}
	// Limit request body size to prevent large memory allocation
	r.Body = nethttp.MaxBytesReader(nil, r.Body, maxBodySize)
	defer r.Body.Close()

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(v); err != nil {
		var maxBytesErr nethttp.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return domain.ErrRequestBodyTooLarge
		}
		return domain.ErrInvalidInput
	}
	return nil
}
