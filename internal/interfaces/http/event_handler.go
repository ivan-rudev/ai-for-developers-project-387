package http

import (
	nethttp "net/http"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/usecase"
)

// EventHandler обрабатывает запросы к событиям владельца.
type EventHandler struct {
	events *usecase.EventUsecase
}

// NewEventHandler создаёт EventHandler.
func NewEventHandler(events *usecase.EventUsecase) *EventHandler {
	return &EventHandler{events: events}
}

// List возвращает активные события владельца.
func (h *EventHandler) List(w nethttp.ResponseWriter, r *nethttp.Request) {
	events, err := h.events.ListActiveByOwner(r.Context(), r.PathValue("uuid"))
	if err != nil {
		writeOwnerNotFound(w, err)
		return
	}
	list := make([]EventResponse, 0, len(events))
	for _, e := range events {
		list = append(list, eventToResponse(e))
	}
	writeJSON(w, nethttp.StatusOK, eventsResponse{Events: list})
}

// Create создаёт событие владельца.
func (h *EventHandler) Create(w nethttp.ResponseWriter, r *nethttp.Request) {
	var req CreateEventRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	event, err := h.events.CreateForOwner(r.Context(), r.PathValue("uuid"), req.Name, req.Description, req.DurationMinutes)
	if err != nil {
		writeOwnerNotFound(w, err)
		return
	}
	writeJSON(w, nethttp.StatusCreated, eventToResponse(*event))
}
