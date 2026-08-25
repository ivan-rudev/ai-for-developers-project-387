// Package fake содержит in-memory реализации репозиториев для юнит-тестов
// use case'ов. Имитируют поведение SQLite-хранилища: unique-constraint'ы
// (email владельца, название события в рамках владельца, email гостя) и
// транзакционную проверку пересечений бронирований.
package fake

import (
	"context"
	"time"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
)

// Repositories — контейнер связанных in-memory fake-репозиториев.
type Repositories struct {
	Owners   *MemoryOwnerRepository
	Guests   *MemoryGuestRepository
	Events   *MemoryEventRepository
	Bookings *MemoryBookingRepository
}

// NewRepositories создаёт пустой контейнер фейков, связанных друг с другом.
func NewRepositories() *Repositories {
	r := &Repositories{}
	r.Owners = NewOwnerRepository()
	r.Guests = NewGuestRepository()
	r.Events = NewEventRepository()
	r.Bookings = NewBookingRepository(r)
	r.Owners.events = r.Events
	return r
}

// AddOwner добавляет владельца в фейк-репозиторий.
func (r *Repositories) AddOwner(o *domain.Owner) *domain.Owner {
	if _, err := r.Owners.Create(context.Background(), o); err != nil {
		panic(err)
	}
	return o
}

// AddEvent добавляет событие в фейк-репозиторий.
func (r *Repositories) AddEvent(e *domain.Event) *domain.Event {
	if _, err := r.Events.Create(context.Background(), e); err != nil {
		panic(err)
	}
	return e
}

// AddGuest добавляет гостя в фейк-репозиторий.
func (r *Repositories) AddGuest(name, email string) *domain.Guest {
	g := &domain.Guest{Name: name, Email: email}
	if _, err := r.Guests.Create(context.Background(), g); err != nil {
		panic(err)
	}
	return g
}

// AddBooking создаёт бронирование в фейк-репозитории и возвращает его.
func (r *Repositories) AddBooking(ownerID, guestID, eventID int64, start, end time.Time) *domain.Booking {
	id, err := r.Bookings.CreateBooking(context.Background(), ownerID, guestID, eventID, start.UTC(), end.UTC())
	if err != nil {
		panic(err)
	}
	b, err := r.Bookings.GetByID(context.Background(), id)
	if err != nil {
		panic(err)
	}
	return b
}
