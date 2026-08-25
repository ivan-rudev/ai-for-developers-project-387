package repository

import (
	"context"
	"time"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
)

// BookingRepository — хранилище бронирований.
type BookingRepository interface {
	GetByOwnerID(ctx context.Context, ownerID int64) ([]domain.Booking, error)
	GetByOwnerIDAndDate(ctx context.Context, ownerID int64, date time.Time) ([]domain.Booking, error)
	GetUpcomingByOwnerID(ctx context.Context, ownerID int64, from time.Time) ([]domain.Booking, error)
	// CreateBooking в одной транзакции проверяет пересечение с существующими
	// бронированиями и вставляет запись (BEGIN IMMEDIATE + UNIQUE constraint).
	CreateBooking(ctx context.Context, ownerID, guestID, eventID int64, start, end time.Time) (int64, error)
}
