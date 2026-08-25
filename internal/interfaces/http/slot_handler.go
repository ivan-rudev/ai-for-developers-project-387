package http

import (
	"context"
	"errors"
	nethttp "net/http"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/interfaces/http/middleware"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/usecase"
)

// SlotHandler обрабатывает запросы к слотам владельца.
type SlotHandler struct {
	owners *usecase.OwnerUsecase
	events *usecase.EventUsecase
	slots  *usecase.SlotUsecase
}

// NewSlotHandler создаёт SlotHandler.
func NewSlotHandler(owners *usecase.OwnerUsecase, events *usecase.EventUsecase, slots *usecase.SlotUsecase) *SlotHandler {
	return &SlotHandler{owners: owners, events: events, slots: slots}
}

// List возвращает слоты события владельца на окно бронирования.
func (h *SlotHandler) List(w nethttp.ResponseWriter, r *nethttp.Request) {
	ctx := r.Context()
	ownerUUID := r.PathValue("uuid")
	eventUUID := r.URL.Query().Get("event_uuid")

	slotsByDate, err := h.slots.GenerateSlots(ctx, ownerUUID, eventUUID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			writeNotFound(w, ctx, h.owners, ownerUUID, "event")
			return
		}
		writeError(w, err)
		return
	}

	owner, err := h.owners.GetByUUID(ctx, ownerUUID)
	if err != nil {
		writeError(w, err)
		return
	}
	event, err := h.events.GetByUUIDForOwner(ctx, ownerUUID, eventUUID)
	if err != nil {
		writeError(w, err)
		return
	}
	startDate, endDate, err := h.slots.BookingWindow(ctx, ownerUUID)
	if err != nil {
		writeError(w, err)
		return
	}

	slots := make(map[string][]SlotResponse, len(slotsByDate))
	for date, daySlots := range slotsByDate {
		items := make([]SlotResponse, 0, len(daySlots))
		for _, s := range daySlots {
			items = append(items, SlotResponse{
				Time:   s.StartTime.Format("15:04"),
				Status: s.Status,
				Reason: s.Reason,
			})
		}
		slots[date] = items
	}

	writeJSON(w, nethttp.StatusOK, SlotsResponse{
		EventUUID:       event.UUID,
		EventName:       event.Name,
		DurationMinutes: event.DurationMinutes,
		Timezone:        owner.Timezone,
		StartDate:       startDate,
		EndDate:         endDate,
		Slots:           slots,
	})
}

// writeNotFound пишет 404 с контекстным сообщением: различает «владелец не
// найден» и «entity не найдено».
func writeNotFound(
	w nethttp.ResponseWriter,
	ctx context.Context,
	owners *usecase.OwnerUsecase,
	ownerUUID string,
	entity string,
) {
	if _, err := owners.GetByUUID(ctx, ownerUUID); errors.Is(err, domain.ErrNotFound) {
		middleware.WriteError(w, nethttp.StatusNotFound, "not_found", "owner not found")
		return
	}
	middleware.WriteError(w, nethttp.StatusNotFound, "not_found", entity+" not found")
}
