package http

import (
	nethttp "net/http"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/usecase"
)

// AdminHandler обрабатывает запросы админской панели default owner.
type AdminHandler struct {
	admin *usecase.AdminUsecase
}

// NewAdminHandler создаёт AdminHandler.
func NewAdminHandler(admin *usecase.AdminUsecase) *AdminHandler {
	return &AdminHandler{admin: admin}
}

// GetOwner возвращает default owner.
func (h *AdminHandler) GetOwner(w nethttp.ResponseWriter, r *nethttp.Request) {
	owner, err := h.admin.GetDefaultOwner(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, nethttp.StatusOK, ownerToResponse(owner))
}

// ListBookings возвращает предстоящие бронирования default owner.
func (h *AdminHandler) ListBookings(w nethttp.ResponseWriter, r *nethttp.Request) {
	bookings, err := h.admin.ListUpcomingBookings(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	list := make([]BookingAdminResponse, 0, len(bookings))
	for _, b := range bookings {
		list = append(list, bookingAdminToResponse(b))
	}
	writeJSON(w, nethttp.StatusOK, bookingsAdminResponse{Bookings: list})
}

// ListEvents возвращает все события default owner, включая неактивные.
func (h *AdminHandler) ListEvents(w nethttp.ResponseWriter, r *nethttp.Request) {
	events, err := h.admin.ListEvents(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	list := make([]EventResponse, 0, len(events))
	for _, e := range events {
		list = append(list, eventToResponse(e))
	}
	writeJSON(w, nethttp.StatusOK, eventsResponse{Events: list})
}

// CreateEvent создаёт событие default owner.
func (h *AdminHandler) CreateEvent(w nethttp.ResponseWriter, r *nethttp.Request) {
	var req CreateEventRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	event, err := h.admin.CreateEvent(r.Context(), req.Name, req.Description, req.DurationMinutes)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, nethttp.StatusCreated, eventToResponse(*event))
}
