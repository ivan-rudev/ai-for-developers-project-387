package domain

import "time"

const (
	// SlotStatusAvailable — слот доступен для бронирования.
	SlotStatusAvailable = "available"
	// SlotStatusUnavailable — слот недоступен.
	SlotStatusUnavailable = "unavailable"
)

const (
	// UnavailableReasonBooked — слот занят бронированием.
	UnavailableReasonBooked = "booked"
	// UnavailableReasonPast — слот уже прошёл.
	UnavailableReasonPast = "past"
)

// Slot — временной интервал для бронирования.
// StartTime и EndTime выражены в локальном времени владельца.
type Slot struct {
	StartTime time.Time
	EndTime   time.Time
	Status    string // "available" | "unavailable"
	Reason    string // "booked" | "past", пусто для available
}
