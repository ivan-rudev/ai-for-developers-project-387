package http

import (
	"errors"
	nethttp "net/http"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/usecase"
)

// BookingHandler обрабатывает запросы к бронированиям.
type BookingHandler struct {
	owners   *usecase.OwnerUsecase
	bookings *usecase.BookingUsecase
}

// NewBookingHandler создаёт BookingHandler.
func NewBookingHandler(owners *usecase.OwnerUsecase, bookings *usecase.BookingUsecase) *BookingHandler {
	return &BookingHandler{owners: owners, bookings: bookings}
}

// Create создаёт бронирование.
func (h *BookingHandler) Create(w nethttp.ResponseWriter, r *nethttp.Request) {
	var req CreateBookingRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, err)
		return
	}

	booking, err := h.bookings.CreateBooking(r.Context(), usecase.CreateBookingInput{
		OwnerUUID:  req.OwnerUUID,
		EventUUID:  req.EventUUID,
		GuestName:  req.GuestName,
		GuestEmail: req.GuestEmail,
		Date:       req.Date,
		StartTime:  req.StartTime,
	})
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeNotFound(w, r.Context(), h.owners, req.OwnerUUID, "event")
			return
		}
		writeError(w, err)
		return
	}

	writeJSON(w, nethttp.StatusCreated, bookingToResponse(req.OwnerUUID, *booking))
}
