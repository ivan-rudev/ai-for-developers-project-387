package usecase_test

import (
	"errors"
	"testing"
	"time"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/usecase"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/usecase/fake"
)

func TestAdminUsecase_GetDefaultOwner(t *testing.T) {
	r, _, _ := fake.SeedBob()
	uc := newAdminUsecase(r, fake.NewFixedClock(fake.Today))

	owner, err := uc.GetDefaultOwner(testCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if owner.Name != "Bob Jones" || owner.UUID != fake.BobUUID {
		t.Fatalf("unexpected owner: %+v", owner)
	}
}

func TestAdminUsecase_ListUpcomingBookings_Sorted(t *testing.T) {
	r, bob, events := fake.SeedBob()
	std := eventByUUID(events, fake.StandardEventUUID)
	guest := r.AddGuest("Alice", "alice@example.com")
	r.AddBooking(bob.ID, guest.ID, std.ID, fake.Moscow(2026, time.August, 12, 10, 0), fake.Moscow(2026, time.August, 12, 10, 30))
	r.AddBooking(bob.ID, guest.ID, std.ID, fake.Moscow(2026, time.August, 11, 10, 0), fake.Moscow(2026, time.August, 11, 10, 30))

	clock := fake.NewFixedClock(fake.Moscow(2026, time.August, 10, 9, 0))
	uc := newAdminUsecase(r, clock)

	bookings, err := uc.ListUpcomingBookings(testCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bookings) != 2 {
		t.Fatalf("expected 2 bookings, got %d", len(bookings))
	}
	if bookings[0].StartTime.Format("01-02") != "08-11" || bookings[1].StartTime.Format("01-02") != "08-12" {
		t.Fatalf("expected sorted by start_time, got %s / %s", bookings[0].StartTime.Format("01-02"), bookings[1].StartTime.Format("01-02"))
	}
	if bookings[0].GuestName != "Alice" || bookings[0].EventName != "Стандартная встреча" || bookings[0].GuestEmail != "alice@example.com" {
		t.Fatalf("unexpected booking fields: %+v", bookings[0])
	}
}

func TestAdminUsecase_ListUpcomingBookings_ExcludesPast(t *testing.T) {
	r, bob, events := fake.SeedBob()
	std := eventByUUID(events, fake.StandardEventUUID)
	guest := r.AddGuest("Alice", "alice@example.com")
	// Прошедшая бронь 2026-08-08.
	r.AddBooking(bob.ID, guest.ID, std.ID, fake.Moscow(2026, time.August, 8, 10, 0), fake.Moscow(2026, time.August, 8, 10, 30))
	// Будущая бронь 2026-08-12.
	r.AddBooking(bob.ID, guest.ID, std.ID, fake.Moscow(2026, time.August, 12, 10, 0), fake.Moscow(2026, time.August, 12, 10, 30))

	clock := fake.NewFixedClock(fake.Moscow(2026, time.August, 10, 9, 0))
	uc := newAdminUsecase(r, clock)

	bookings, err := uc.ListUpcomingBookings(testCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bookings) != 1 || bookings[0].StartTime.Format("01-02") != "08-12" {
		t.Fatalf("expected only future booking, got %#v", bookings)
	}
}

func TestAdminUsecase_ListUpcomingBookings_Empty(t *testing.T) {
	r := fake.NewRepositories()
	fake.SeedBobInto(r)
	clock := fake.NewFixedClock(fake.Moscow(2026, time.August, 10, 9, 0))
	uc := newAdminUsecase(r, clock)

	bookings, err := uc.ListUpcomingBookings(testCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bookings == nil || len(bookings) != 0 {
		t.Fatalf("expected empty list, got %#v", bookings)
	}
}

func TestAdminUsecase_ListEvents_IncludesInactive(t *testing.T) {
	r, bob, _ := fake.SeedBob()
	r.AddEvent(&domain.Event{UUID: "6ba7b810-9dad-41d4-a716-446655444000", OwnerID: bob.ID, Name: "Консультация", DurationMinutes: 45, IsActive: false})
	uc := newAdminUsecase(r, fake.NewFixedClock(fake.Today))

	events, err := uc.ListEvents(testCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events (incl. inactive), got %d", len(events))
	}
}

func TestAdminUsecase_CreateEvent_HappyPath(t *testing.T) {
	r, _, _ := fake.SeedBob()
	uc := newAdminUsecase(r, fake.NewFixedClock(fake.Today))

	ev, err := uc.CreateEvent(testCtx, "Длинная встреча", "60 минут", 60)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isUUIDv4(ev.UUID) || ev.Name != "Длинная встреча" || ev.DurationMinutes != 60 || !ev.IsActive {
		t.Fatalf("unexpected event: %+v", ev)
	}
}

func TestAdminUsecase_CreateEvent_DuplicateName(t *testing.T) {
	r, _, _ := fake.SeedBob()
	uc := newAdminUsecase(r, fake.NewFixedClock(fake.Today))

	_, err := uc.CreateEvent(testCtx, "Короткая встреча", "", 15)
	if !errors.Is(err, usecase.ErrEventNameExists) {
		t.Fatalf("expected ErrEventNameExists, got %v", err)
	}
}

func TestAdminUsecase_CreateEvent_InvalidDuration(t *testing.T) {
	r, _, _ := fake.SeedBob()
	uc := newAdminUsecase(r, fake.NewFixedClock(fake.Today))

	_, err := uc.CreateEvent(testCtx, "Событие", "", -5)
	if !errors.Is(err, usecase.ErrInvalidDuration) {
		t.Fatalf("expected ErrInvalidDuration, got %v", err)
	}
}
