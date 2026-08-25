package domain

import "time"

// Booking — бронирование временного слота. Содержит денормализованные поля
// для удобного отображения (имя/email гостя, название/длительность события).
type Booking struct {
	ID              int64
	OwnerID         int64
	GuestID         int64
	EventID         int64
	GuestName       string
	EventName       string
	GuestEmail      string
	DurationMinutes int
	StartTime       time.Time // UTC
	EndTime         time.Time // UTC
	CreatedAt       time.Time
}
