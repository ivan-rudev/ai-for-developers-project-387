package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
)

// BookingRepository — реализация repository.BookingRepository на SQLite.
type BookingRepository struct {
	db *DB
}

// NewBookingRepository создаёт репозиторий бронирований.
func NewBookingRepository(db *DB) *BookingRepository {
	return &BookingRepository{db: db}
}

// bookingSelect — выборка бронирований с денормализованными полями гостя и события.
const bookingSelect = `
SELECT b.id, b.owner_id, b.guest_id, b.event_id,
       g.name, e.name, g.email, e.duration_minutes,
       b.start_time, b.end_time, b.created_at
FROM bookings b
JOIN guests g ON g.id = b.guest_id
JOIN events e ON e.id = b.event_id
`

// GetByOwnerID возвращает все бронирования владельца, отсортированные по start_time.
func (r *BookingRepository) GetByOwnerID(ctx context.Context, ownerID int64) ([]domain.Booking, error) {
	rows, err := r.db.Conn().QueryContext(ctx,
		bookingSelect+` WHERE b.owner_id = ? ORDER BY b.start_time`, ownerID)
	if err != nil {
		return nil, fmt.Errorf("list bookings: %w", err)
	}
	defer rows.Close()

	return collectBookings(rows)
}

// GetByOwnerIDAndDate возвращает бронирования владельца, начинающиеся в течение
// суток (UTC) даты date: start_time ∈ [date, date+24h).
func (r *BookingRepository) GetByOwnerIDAndDate(ctx context.Context, ownerID int64, date time.Time) ([]domain.Booking, error) {
	dateUTC := date.UTC().Truncate(24 * time.Hour)
	end := dateUTC.Add(24 * time.Hour)

	rows, err := r.db.Conn().QueryContext(ctx,
		bookingSelect+` WHERE b.owner_id = ? AND b.start_time >= ? AND b.start_time < ? ORDER BY b.start_time`,
		ownerID, utcFormat(dateUTC), utcFormat(end))
	if err != nil {
		return nil, fmt.Errorf("list bookings by date: %w", err)
	}
	defer rows.Close()

	return collectBookings(rows)
}

// GetUpcomingByOwnerID возвращает бронирования владельца с start_time >= from,
// отсортированные по start_time.
func (r *BookingRepository) GetUpcomingByOwnerID(ctx context.Context, ownerID int64, from time.Time) ([]domain.Booking, error) {
	rows, err := r.db.Conn().QueryContext(ctx,
		bookingSelect+` WHERE b.owner_id = ? AND b.start_time >= ? ORDER BY b.start_time`,
		ownerID, utcFormat(from))
	if err != nil {
		return nil, fmt.Errorf("list upcoming bookings: %w", err)
	}
	defer rows.Close()

	return collectBookings(rows)
}

// CreateBooking в одной транзакции (BEGIN IMMEDIATE) проверяет пересечение с
// существующими бронированиями и вставляет запись. Гарантия отсутствия гонок
// дублируется UNIQUE(owner_id, start_time, end_time).
//
// Возвращает domain.ErrOverlap при пересечении с существующей бронью.
func (r *BookingRepository) CreateBooking(ctx context.Context, ownerID, guestID, eventID int64, start, end time.Time) (int64, error) {
	var bookingID int64

	err := r.db.withinTx(ctx, func(tx *sql.Tx) error {
		var overlaps int
		if err := tx.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM bookings WHERE owner_id = ? AND start_time < ? AND end_time > ?`,
			ownerID, utcFormat(end), utcFormat(start),
		).Scan(&overlaps); err != nil {
			return fmt.Errorf("check booking overlap: %w", err)
		}
		if overlaps > 0 {
			return domain.ErrOverlap
		}

		res, err := tx.ExecContext(ctx, `
			INSERT INTO bookings (owner_id, guest_id, event_id, start_time, end_time, created_at)
			VALUES (?, ?, ?, ?, ?, ?)`,
			ownerID, guestID, eventID, utcFormat(start), utcFormat(end), utcFormat(time.Now()),
		)
		if err != nil {
			return mapStoreError(err)
		}

		bookingID, err = res.LastInsertId()
		if err != nil {
			return fmt.Errorf("last insert id: %w", err)
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	return bookingID, nil
}

func collectBookings(rows *sql.Rows) ([]domain.Booking, error) {
	bookings := make([]domain.Booking, 0)
	for rows.Next() {
		booking, err := scanBooking(rows)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bookings: %w", err)
	}
	return bookings, nil
}

func scanBooking(s scanner) (domain.Booking, error) {
	var (
		booking      domain.Booking
		startRaw     string
		endRaw       string
		createdRaw   string
		guestName    string
		eventName    string
		guestEmail   string
		durationMins int
	)
	if err := s.Scan(
		&booking.ID,
		&booking.OwnerID,
		&booking.GuestID,
		&booking.EventID,
		&guestName,
		&eventName,
		&guestEmail,
		&durationMins,
		&startRaw,
		&endRaw,
		&createdRaw,
	); err != nil {
		return domain.Booking{}, err
	}

	booking.GuestName = guestName
	booking.EventName = eventName
	booking.GuestEmail = guestEmail
	booking.DurationMinutes = durationMins

	var err error
	if booking.StartTime, err = sqlTimeParse(startRaw); err != nil {
		return domain.Booking{}, err
	}
	if booking.EndTime, err = sqlTimeParse(endRaw); err != nil {
		return domain.Booking{}, err
	}
	if booking.CreatedAt, err = sqlTimeParse(createdRaw); err != nil {
		return domain.Booking{}, err
	}

	return booking, nil
}
