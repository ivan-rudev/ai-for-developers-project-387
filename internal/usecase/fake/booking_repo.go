package fake

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
)

// MemoryBookingRepository — in-memory BookingRepository. CreateBooking имитирует
// транзакцию с UNIQUE(owner_id, start_time, end_time): пересечение с любой
// существующей бронью владельца (полуинтервалы) → domain.ErrOverlap.
type MemoryBookingRepository struct {
	mu     sync.Mutex
	nextID int64
	list   []*domain.Booking
	byID   map[int64]*domain.Booking
	stores *Repositories
}

func NewBookingRepository(stores *Repositories) *MemoryBookingRepository {
	return &MemoryBookingRepository{
		byID:   make(map[int64]*domain.Booking),
		stores: stores,
	}
}

// GetByOwnerID возвращает бронирования владельца, отсортированные по start_time.
func (r *MemoryBookingRepository) GetByOwnerID(_ context.Context, ownerID int64) ([]domain.Booking, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]domain.Booking, 0)
	for _, b := range r.list {
		if b.OwnerID == ownerID {
			out = append(out, *b)
		}
	}
	sortBookingsByStart(out)
	return out, nil
}

// GetByOwnerIDAndDate возвращает бронирования владельца за календарную дату (UTC).
func (r *MemoryBookingRepository) GetByOwnerIDAndDate(_ context.Context, ownerID int64, date time.Time) ([]domain.Booking, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.AddDate(0, 0, 1)
	out := make([]domain.Booking, 0)
	for _, b := range r.list {
		if b.OwnerID == ownerID && !b.StartTime.Before(dayStart) && b.StartTime.Before(dayEnd) {
			out = append(out, *b)
		}
	}
	sortBookingsByStart(out)
	return out, nil
}

// GetUpcomingByOwnerID возвращает бронирования с start_time >= from, отсортированные по start_time.
func (r *MemoryBookingRepository) GetUpcomingByOwnerID(_ context.Context, ownerID int64, from time.Time) ([]domain.Booking, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]domain.Booking, 0)
	for _, b := range r.list {
		if b.OwnerID == ownerID && !b.StartTime.Before(from) {
			out = append(out, *b)
		}
	}
	sortBookingsByStart(out)
	return out, nil
}

func (r *MemoryBookingRepository) GetByID(_ context.Context, id int64) (*domain.Booking, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if b, ok := r.byID[id]; ok {
		cp := *b
		return &cp, nil
	}
	return nil, domain.ErrNotFound
}

func (r *MemoryBookingRepository) CreateBooking(_ context.Context, ownerID, guestID, eventID int64, start, end time.Time) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, b := range r.list {
		if b.OwnerID == ownerID && start.Before(b.EndTime) && end.After(b.StartTime) {
			return 0, domain.ErrOverlap
		}
	}
	guest := r.guestByID(guestID)
	event := r.eventByID(eventID)
	r.nextID++
	b := &domain.Booking{
		ID:        r.nextID,
		OwnerID:   ownerID,
		GuestID:   guestID,
		EventID:   eventID,
		StartTime: start.UTC(),
		EndTime:   end.UTC(),
		CreatedAt: time.Now().UTC(),
	}
	if guest != nil {
		b.GuestName = guest.Name
		b.GuestEmail = guest.Email
	}
	if event != nil {
		b.EventName = event.Name
		b.DurationMinutes = event.DurationMinutes
	}
	r.list = append(r.list, b)
	r.byID[b.ID] = b
	return b.ID, nil
}

// Count возвращает число бронирований.
func (r *MemoryBookingRepository) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.list)
}

func (r *MemoryBookingRepository) guestByID(id int64) *domain.Guest {
	r.stores.Guests.mu.Lock()
	defer r.stores.Guests.mu.Unlock()
	return r.stores.Guests.byID[id]
}

func (r *MemoryBookingRepository) eventByID(id int64) *domain.Event {
	r.stores.Events.mu.Lock()
	defer r.stores.Events.mu.Unlock()
	return r.stores.Events.byID[id]
}

func sortBookingsByStart(bookings []domain.Booking) {
	sort.Slice(bookings, func(i, j int) bool {
		return bookings[i].StartTime.Before(bookings[j].StartTime)
	})
}
