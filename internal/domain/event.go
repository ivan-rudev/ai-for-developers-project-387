package domain

import "time"

// Event — тип события владельца. Определяет длительность встречи
// и заменяет собой уровень slot_duration у владельца.
type Event struct {
	ID              int64
	UUID            string
	OwnerID         int64
	Name            string
	Description     string
	DurationMinutes int
	IsActive        bool
	CreatedAt       time.Time
}
