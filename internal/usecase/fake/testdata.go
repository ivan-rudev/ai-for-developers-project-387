package fake

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
)

// UUID seed-владельца и seed-событий (docs/TESTING.md §0.1–0.2).
const (
	BobUUID           = "550e8400-e29b-41d4-a716-446655440000"
	ShortEventUUID    = "6ba7b810-9dad-41d4-a716-446655440001"
	StandardEventUUID = "6ba7b810-9dad-41d4-a716-446655440002"
)

// Today — фиксированная «сегодняшняя» дата тестов: 2026-08-10 (понедельник).
var Today = time.Date(2026, time.August, 10, 0, 0, 0, 0, time.UTC)

// SeedBob создаёт фейк-репозитории с seed-владельцем Bob и его двумя событиями.
func SeedBob() (*Repositories, *domain.Owner, []domain.Event) {
	r := NewRepositories()
	bob, events := SeedBobInto(r)
	return r, bob, events
}

// SeedBobInto заполняет фейк-репозитории seed-владельцем Bob и его событиями.
func SeedBobInto(r *Repositories) (*domain.Owner, []domain.Event) {
	ctx := context.Background()

	bob := NewMoscowOwner(BobUUID, "Bob Jones", "bob@example.com")
	if _, err := r.Owners.Create(ctx, bob); err != nil {
		panic(err)
	}

	short := &domain.Event{UUID: ShortEventUUID, OwnerID: bob.ID, Name: "Короткая встреча", DurationMinutes: 15, IsActive: true}
	std := &domain.Event{UUID: StandardEventUUID, OwnerID: bob.ID, Name: "Стандартная встреча", DurationMinutes: 30, IsActive: true}
	if _, err := r.Events.Create(ctx, short); err != nil {
		panic(err)
	}
	if _, err := r.Events.Create(ctx, std); err != nil {
		panic(err)
	}
	return bob, []domain.Event{*short, *std}
}

// NewMoscowOwner создаёт владельца с настройками seed-владельца
// (Europe/Moscow, 09:00–18:00, пн–пт).
func NewMoscowOwner(uuid, name, email string) *domain.Owner {
	return &domain.Owner{
		UUID:        uuid,
		Name:        name,
		Email:       email,
		IsActive:    true,
		WorkStart:   "09:00",
		WorkEnd:     "18:00",
		Timezone:    "Europe/Moscow",
		WorkingDays: weekdaysMonFri(),
	}
}

// NewNYOwner создаёт владельца в часовом поясе America/New_York
// (09:00–17:00, пн–пт).
func NewNYOwner(uuid, name, email string) *domain.Owner {
	return &domain.Owner{
		UUID:        uuid,
		Name:        name,
		Email:       email,
		IsActive:    true,
		WorkStart:   "09:00",
		WorkEnd:     "17:00",
		Timezone:    "America/New_York",
		WorkingDays: weekdaysMonFri(),
	}
}

// AllWorkingDays возвращает рабочие дни «всю неделю».
func AllWorkingDays() map[string]bool {
	return map[string]bool{
		"mon": true, "tue": true, "wed": true, "thu": true,
		"fri": true, "sat": true, "sun": true,
	}
}

// FixedClock — фиксированный источник времени для детерминированных тестов.
type FixedClock struct {
	now time.Time
}

// NewFixedClock создаёт FixedClock с заданным моментом времени.
func NewFixedClock(at time.Time) *FixedClock { return &FixedClock{now: at} }

// Now возвращает фиксированный момент времени.
func (c *FixedClock) Now() time.Time { return c.now }

// StubUUID — детерминированный генератор валидных UUID v4 для тестов.
type StubUUID struct {
	mu sync.Mutex
	n  int
}

// NewStubUUID создаёт StubUUID, нумерующий UUID с start+1.
func NewStubUUID(start int) *StubUUID { return &StubUUID{n: start} }

// New возвращает следующий детерминированный UUID v4.
func (s *StubUUID) New() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.n++
	return fmt.Sprintf("00000000-0000-4000-8000-%012d", s.n), nil
}

// Moscow возвращает локальное время в Europe/Moscow.
func Moscow(y int, m time.Month, d, hour, min int) time.Time {
	return time.Date(y, m, d, hour, min, 0, 0, mustLoc("Europe/Moscow"))
}

// NY возвращает локальное время в America/New_York.
func NY(y int, m time.Month, d, hour, min int) time.Time {
	return time.Date(y, m, d, hour, min, 0, 0, mustLoc("America/New_York"))
}

func weekdaysMonFri() map[string]bool {
	return map[string]bool{"mon": true, "tue": true, "wed": true, "thu": true, "fri": true}
}

func mustLoc(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}
