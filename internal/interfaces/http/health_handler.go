package http

import (
	nethttp "net/http"
)

// HealthHandler обрабатывает healthcheck-запросы.
type HealthHandler struct{}

// NewHealthHandler создаёт HealthHandler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Check возвращает статус приложения.
func (h *HealthHandler) Check(w nethttp.ResponseWriter, _ *nethttp.Request) {
	writeJSON(w, nethttp.StatusOK, struct {
		Status string `json:"status"`
	}{Status: "ok"})
}
