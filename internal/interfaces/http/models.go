package http

import (
	"time"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
)

// OwnerResponse — владелец с настройками календаря.
type OwnerResponse struct {
	UUID     string        `json:"uuid"`
	Name     string        `json:"name"`
	Settings OwnerSettings `json:"settings"`
}

// OwnerSettings — настройки календаря владельца.
type OwnerSettings struct {
	WorkStart   string   `json:"work_start"`
	WorkEnd     string   `json:"work_end"`
	Timezone    string   `json:"timezone"`
	WorkingDays []string `json:"working_days"`
}

// OwnerSummary — краткая информация о владельце.
type OwnerSummary struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
}

// EventResponse — событие владельца.
type EventResponse struct {
	UUID            string `json:"uuid"`
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	DurationMinutes int    `json:"duration_minutes"`
	IsActive        bool   `json:"is_active"`
}

// BookingResponse — бронирование публичного API.
type BookingResponse struct {
	OwnerUUID       string `json:"owner_uuid"`
	GuestName       string `json:"guest_name"`
	EventName       string `json:"event_name"`
	DurationMinutes int    `json:"duration_minutes"`
	StartTime       string `json:"start_time"`
	EndTime         string `json:"end_time"`
	CreatedAt       string `json:"created_at,omitempty"`
}

// BookingAdminResponse — бронирование админской панели.
type BookingAdminResponse struct {
	EventName       string `json:"event_name"`
	DurationMinutes int    `json:"duration_minutes"`
	GuestName       string `json:"guest_name"`
	GuestEmail      string `json:"guest_email"`
	StartTime       string `json:"start_time"`
	EndTime         string `json:"end_time"`
	CreatedAt       string `json:"created_at,omitempty"`
}

// SlotResponse — слот события.
type SlotResponse struct {
	Time   string `json:"time"`
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

// SlotsResponse — слоты события на окно бронирования, сгруппированные по дате.
type SlotsResponse struct {
	EventUUID       string                    `json:"event_uuid"`
	EventName       string                    `json:"event_name"`
	DurationMinutes int                       `json:"duration_minutes"`
	Timezone        string                    `json:"timezone"`
	StartDate       string                    `json:"start_date"`
	EndDate         string                    `json:"end_date"`
	Slots           map[string][]SlotResponse `json:"slots"`
}

// CreateOwnerRequest — тело запроса на создание владельца.
type CreateOwnerRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// CreateEventRequest — тело запроса на создание события.
type CreateEventRequest struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	DurationMinutes int    `json:"duration_minutes"`
}

// CreateBookingRequest — тело запроса на создание бронирования.
type CreateBookingRequest struct {
	OwnerUUID  string `json:"owner_uuid"`
	EventUUID  string `json:"event_uuid"`
	GuestName  string `json:"guest_name"`
	GuestEmail string `json:"guest_email"`
	Date       string `json:"date"`
	StartTime  string `json:"start_time"`
}

type bookingsResponse struct {
	Bookings []BookingResponse `json:"bookings"`
}

type bookingsAdminResponse struct {
	Bookings []BookingAdminResponse `json:"bookings"`
}

type eventsResponse struct {
	Events []EventResponse `json:"events"`
}

// ownerToResponse преобразует доменного владельца в OwnerResponse.
func ownerToResponse(o *domain.Owner) OwnerResponse {
	return OwnerResponse{
		UUID: o.UUID,
		Name: o.Name,
		Settings: OwnerSettings{
			WorkStart:   o.WorkStart,
			WorkEnd:     o.WorkEnd,
			Timezone:    o.Timezone,
			WorkingDays: orderedWorkingDays(o.WorkingDays),
		},
	}
}

// orderedWorkingDays возвращает рабочие дни в фиксированном порядке пн–вс.
func orderedWorkingDays(days map[string]bool) []string {
	out := make([]string, 0, len(days))
	for _, d := range workingDayOrder {
		if days[d] {
			out = append(out, d)
		}
	}
	return out
}

var workingDayOrder = []string{"mon", "tue", "wed", "thu", "fri", "sat", "sun"}

func eventToResponse(e domain.Event) EventResponse {
	return EventResponse{
		UUID:            e.UUID,
		Name:            e.Name,
		Description:     e.Description,
		DurationMinutes: e.DurationMinutes,
		IsActive:        e.IsActive,
	}
}

func bookingToResponse(ownerUUID string, b domain.Booking) BookingResponse {
	return BookingResponse{
		OwnerUUID:       ownerUUID,
		GuestName:       b.GuestName,
		EventName:       b.EventName,
		DurationMinutes: b.DurationMinutes,
		StartTime:       b.StartTime.UTC().Format(time.RFC3339),
		EndTime:         b.EndTime.UTC().Format(time.RFC3339),
		CreatedAt:       formatTime(b.CreatedAt),
	}
}

func bookingAdminToResponse(b domain.Booking) BookingAdminResponse {
	return BookingAdminResponse{
		EventName:       b.EventName,
		DurationMinutes: b.DurationMinutes,
		GuestName:       b.GuestName,
		GuestEmail:      b.GuestEmail,
		StartTime:       b.StartTime.UTC().Format(time.RFC3339),
		EndTime:         b.EndTime.UTC().Format(time.RFC3339),
		CreatedAt:       formatTime(b.CreatedAt),
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
