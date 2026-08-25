package http

import (
	nethttp "net/http"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/usecase"
)

// OwnerHandler обрабатывает запросы к владельцам календарей.
type OwnerHandler struct {
	owners *usecase.OwnerUsecase
}

// NewOwnerHandler создаёт OwnerHandler.
func NewOwnerHandler(owners *usecase.OwnerUsecase) *OwnerHandler {
	return &OwnerHandler{owners: owners}
}

// List возвращает список активных владельцев.
func (h *OwnerHandler) List(w nethttp.ResponseWriter, r *nethttp.Request) {
	owners, err := h.owners.ListActiveOwners(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	list := make([]OwnerSummary, 0, len(owners))
	for _, o := range owners {
		list = append(list, OwnerSummary{UUID: o.UUID, Name: o.Name})
	}
	writeJSON(w, nethttp.StatusOK, list)
}

// Get возвращает владельца по UUID.
func (h *OwnerHandler) Get(w nethttp.ResponseWriter, r *nethttp.Request) {
	owner, err := h.owners.GetByUUID(r.Context(), r.PathValue("uuid"))
	if err != nil {
		writeOwnerNotFound(w, err)
		return
	}
	writeJSON(w, nethttp.StatusOK, ownerToResponse(owner))
}

// Create создаёт владельца с default events.
func (h *OwnerHandler) Create(w nethttp.ResponseWriter, r *nethttp.Request) {
	var req CreateOwnerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}
	owner, err := h.owners.CreateOwner(r.Context(), req.Name, req.Email)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, nethttp.StatusCreated, ownerToResponse(owner))
}

// ListBookings возвращает бронирования владельца.
func (h *OwnerHandler) ListBookings(w nethttp.ResponseWriter, r *nethttp.Request) {
	ownerUUID := r.PathValue("uuid")
	bookings, err := h.owners.ListBookings(r.Context(), ownerUUID)
	if err != nil {
		writeOwnerNotFound(w, err)
		return
	}
	list := make([]BookingResponse, 0, len(bookings))
	for _, b := range bookings {
		list = append(list, bookingToResponse(ownerUUID, b))
	}
	writeJSON(w, nethttp.StatusOK, bookingsResponse{Bookings: list})
}
