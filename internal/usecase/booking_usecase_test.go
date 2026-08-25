package usecase_test

import (
	"errors"
	"testing"
	"time"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/usecase"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/usecase/fake"
)

func bookingInput(ownerUUID, eventUUID, guestName, guestEmail, date, startTime string) usecase.CreateBookingInput {
	return usecase.CreateBookingInput{
		OwnerUUID:  ownerUUID,
		EventUUID:  eventUUID,
		GuestName:  guestName,
		GuestEmail: guestEmail,
		Date:       date,
		StartTime:  startTime,
	}
}

func TestBookingUsecase_CreateBooking_HappyPath(t *testing.T) {
	r, bob, _ := fake.SeedBob()
	uc := newBookingUsecase(r, fake.NewFixedClock(fake.Today))

	b, err := uc.CreateBooking(testCtx, bookingInput(bob.UUID, fake.StandardEventUUID, "Alice", "alice@example.com", "2026-08-12", "10:00"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.StartTime.Format(time.RFC3339) != "2026-08-12T07:00:00Z" {
		t.Fatalf("unexpected start_time: %s", b.StartTime.Format(time.RFC3339))
	}
	if b.EndTime.Format(time.RFC3339) != "2026-08-12T07:30:00Z" {
		t.Fatalf("unexpected end_time: %s", b.EndTime.Format(time.RFC3339))
	}
	if b.EventName != "Стандартная встреча" || b.DurationMinutes != 30 {
		t.Fatalf("unexpected event fields: %+v", b)
	}
	if b.GuestName != "Alice" || b.GuestEmail != "alice@example.com" {
		t.Fatalf("unexpected guest fields: %+v", b)
	}
	if b.CreatedAt.IsZero() {
		t.Fatal("created_at must be set")
	}
	if _, err := r.Guests.GetByEmail(testCtx, "alice@example.com"); err != nil {
		t.Fatalf("guest must be created: %v", err)
	}
}

func TestBookingUsecase_CreateBooking_ReusesGuest(t *testing.T) {
	r, bob, _ := fake.SeedBob()
	uc := newBookingUsecase(r, fake.NewFixedClock(fake.Today))

	first, err := uc.CreateBooking(testCtx, bookingInput(bob.UUID, fake.StandardEventUUID, "Alice", "alice@example.com", "2026-08-12", "10:00"))
	if err != nil {
		t.Fatal(err)
	}
	second, err := uc.CreateBooking(testCtx, bookingInput(bob.UUID, fake.ShortEventUUID, "Alice", "alice@example.com", "2026-08-12", "11:00"))
	if err != nil {
		t.Fatal(err)
	}
	if r.Guests.Count() != 1 {
		t.Fatalf("expected 1 guest, got %d", r.Guests.Count())
	}
	if first.GuestID != second.GuestID {
		t.Fatalf("expected same guest reused, got %d != %d", first.GuestID, second.GuestID)
	}
}

func TestBookingUsecase_CreateBooking_SlotTaken(t *testing.T) {
	r, bob, _ := fake.SeedBob()
	uc := newBookingUsecase(r, fake.NewFixedClock(fake.Today))

	if _, err := uc.CreateBooking(testCtx, bookingInput(bob.UUID, fake.StandardEventUUID, "Alice", "alice@example.com", "2026-08-12", "10:00")); err != nil {
		t.Fatal(err)
	}
	_, err := uc.CreateBooking(testCtx, bookingInput(bob.UUID, fake.StandardEventUUID, "Bob Clone", "bobclone@example.com", "2026-08-12", "10:00"))
	if !errors.Is(err, usecase.ErrOverlap) {
		t.Fatalf("expected ErrOverlap, got %v", err)
	}
	if r.Bookings.Count() != 1 {
		t.Fatalf("expected 1 booking, got %d", r.Bookings.Count())
	}
}

func TestBookingUsecase_CreateBooking_PartialOverlap(t *testing.T) {
	r, bob, _ := fake.SeedBob()
	uc := newBookingUsecase(r, fake.NewFixedClock(fake.Today))

	if _, err := uc.CreateBooking(testCtx, bookingInput(bob.UUID, fake.StandardEventUUID, "Alice", "alice@example.com", "2026-08-12", "10:00")); err != nil {
		t.Fatal(err)
	}
	// 15-минутное событие, начинающееся в середине занятого слота 10:00-10:30.
	_, err := uc.CreateBooking(testCtx, bookingInput(bob.UUID, fake.ShortEventUUID, "Carol", "carol@example.com", "2026-08-12", "10:15"))
	if !errors.Is(err, usecase.ErrOverlap) {
		t.Fatalf("expected ErrOverlap, got %v", err)
	}
}

func TestBookingUsecase_CreateBooking_EndAtWorkEnd(t *testing.T) {
	r, bob, _ := fake.SeedBob()
	uc := newBookingUsecase(r, fake.NewFixedClock(fake.Today))

	b, err := uc.CreateBooking(testCtx, bookingInput(bob.UUID, fake.StandardEventUUID, "Alice", "alice@example.com", "2026-08-12", "17:30"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.StartTime.Format(time.RFC3339) != "2026-08-12T14:30:00Z" || b.EndTime.Format(time.RFC3339) != "2026-08-12T15:00:00Z" {
		t.Fatalf("unexpected times: %s..%s", b.StartTime.Format(time.RFC3339), b.EndTime.Format(time.RFC3339))
	}
}

func TestBookingUsecase_CreateBooking_LastDayOfWindow(t *testing.T) {
	r, bob, _ := fake.SeedBob()
	uc := newBookingUsecase(r, fake.NewFixedClock(fake.Today))

	// 2026-08-21 (пятница) — последний рабочий день окна.
	b, err := uc.CreateBooking(testCtx, bookingInput(bob.UUID, fake.StandardEventUUID, "Alice", "alice@example.com", "2026-08-21", "10:00"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.StartTime.Format("2006-01-02") != "2026-08-21" {
		t.Fatalf("unexpected date: %s", b.StartTime.Format("2006-01-02"))
	}
}

func TestBookingUsecase_CreateBooking_DateOutOfRange(t *testing.T) {
	r, bob, _ := fake.SeedBob()
	uc := newBookingUsecase(r, fake.NewFixedClock(fake.Today))

	// 2026-08-24 — за пределами окна 10..23 августа.
	_, err := uc.CreateBooking(testCtx, bookingInput(bob.UUID, fake.StandardEventUUID, "Alice", "alice@example.com", "2026-08-24", "10:00"))
	if !errors.Is(err, usecase.ErrDateOutOfRange) {
		t.Fatalf("expected ErrDateOutOfRange, got %v", err)
	}
	if r.Bookings.Count() != 0 {
		t.Fatal("no booking must be created")
	}
}

func TestBookingUsecase_CreateBooking_Weekend(t *testing.T) {
	r, bob, _ := fake.SeedBob()
	uc := newBookingUsecase(r, fake.NewFixedClock(fake.Today))

	// 2026-08-15 — суббота.
	_, err := uc.CreateBooking(testCtx, bookingInput(bob.UUID, fake.StandardEventUUID, "Alice", "alice@example.com", "2026-08-15", "10:00"))
	if !errors.Is(err, usecase.ErrNotWorkingDay) {
		t.Fatalf("expected ErrNotWorkingDay, got %v", err)
	}
}

func TestBookingUsecase_CreateBooking_OutsideWorkingHours(t *testing.T) {
	r, bob, _ := fake.SeedBob()
	uc := newBookingUsecase(r, fake.NewFixedClock(fake.Today))

	_, err := uc.CreateBooking(testCtx, bookingInput(bob.UUID, fake.StandardEventUUID, "Alice", "alice@example.com", "2026-08-12", "08:30"))
	if !errors.Is(err, usecase.ErrSlotOutsideHours) {
		t.Fatalf("expected ErrSlotOutsideHours, got %v", err)
	}
}

func TestBookingUsecase_CreateBooking_NotMultipleOfDuration(t *testing.T) {
	r, bob, _ := fake.SeedBob()
	uc := newBookingUsecase(r, fake.NewFixedClock(fake.Today))

	// 10:10 не кратно 30 минутам.
	_, err := uc.CreateBooking(testCtx, bookingInput(bob.UUID, fake.StandardEventUUID, "Alice", "alice@example.com", "2026-08-12", "10:10"))
	if !errors.Is(err, usecase.ErrSlotNotMultiple) {
		t.Fatalf("expected ErrSlotNotMultiple, got %v", err)
	}
}

func TestBookingUsecase_CreateBooking_InPast(t *testing.T) {
	r, bob, _ := fake.SeedBob()
	// now = 2026-08-10 14:00.
	clock := fake.NewFixedClock(fake.Moscow(2026, time.August, 10, 14, 0))
	uc := newBookingUsecase(r, clock)

	_, err := uc.CreateBooking(testCtx, bookingInput(bob.UUID, fake.StandardEventUUID, "Alice", "alice@example.com", "2026-08-10", "13:30"))
	if !errors.Is(err, usecase.ErrSlotInPast) {
		t.Fatalf("expected ErrSlotInPast, got %v", err)
	}
}

func TestBookingUsecase_CreateBooking_InvalidGuestEmail(t *testing.T) {
	r, bob, _ := fake.SeedBob()
	uc := newBookingUsecase(r, fake.NewFixedClock(fake.Today))

	_, err := uc.CreateBooking(testCtx, bookingInput(bob.UUID, fake.StandardEventUUID, "Alice", "not-an-email", "2026-08-12", "10:00"))
	if !errors.Is(err, usecase.ErrInvalidEmail) {
		t.Fatalf("expected ErrInvalidEmail, got %v", err)
	}
	if r.Bookings.Count() != 0 {
		t.Fatal("no booking must be created")
	}
}

func TestBookingUsecase_CreateBooking_OwnerNotFound(t *testing.T) {
	r, _, _ := fake.SeedBob()
	uc := newBookingUsecase(r, fake.NewFixedClock(fake.Today))

	_, err := uc.CreateBooking(testCtx, bookingInput("00000000-0000-4000-8000-000000000000", fake.StandardEventUUID, "Alice", "alice@example.com", "2026-08-12", "10:00"))
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestBookingUsecase_CreateBooking_EventNotOwnedOrInactive(t *testing.T) {
	r, bob, _ := fake.SeedBob()
	alice := fake.NewMoscowOwner("6ba7b810-9dad-41d4-a716-446655443800", "Alice Smith", "alice@example.com")
	r.AddOwner(alice)
	aliceEvent := r.AddEvent(&domain.Event{UUID: "6ba7b810-9dad-41d4-a716-446655443801", OwnerID: alice.ID, Name: "Её встреча", DurationMinutes: 30, IsActive: true})
	uc := newBookingUsecase(r, fake.NewFixedClock(fake.Today))

	// Событие другого владельца.
	_, err := uc.CreateBooking(testCtx, bookingInput(bob.UUID, aliceEvent.UUID, "Alice", "alice@example.com", "2026-08-12", "10:00"))
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	// Неактивное событие.
	inactive := r.AddEvent(&domain.Event{UUID: "6ba7b810-9dad-41d4-a716-446655443802", OwnerID: bob.ID, Name: "Скрытое", DurationMinutes: 30, IsActive: false})
	_, err = uc.CreateBooking(testCtx, bookingInput(bob.UUID, inactive.UUID, "Alice", "alice@example.com", "2026-08-12", "10:00"))
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if r.Bookings.Count() != 0 {
		t.Fatal("no booking must be created")
	}
}

func TestBookingUsecase_CreateBooking_InvalidDateAndTime(t *testing.T) {
	r, bob, _ := fake.SeedBob()
	uc := newBookingUsecase(r, fake.NewFixedClock(fake.Today))

	_, err := uc.CreateBooking(testCtx, bookingInput(bob.UUID, fake.StandardEventUUID, "Alice", "alice@example.com", "2026/08/12", "10:00"))
	if !errors.Is(err, usecase.ErrInvalidDate) {
		t.Fatalf("expected ErrInvalidDate, got %v", err)
	}
	_, err = uc.CreateBooking(testCtx, bookingInput(bob.UUID, fake.StandardEventUUID, "Alice", "alice@example.com", "2026-08-12", "25:99"))
	if !errors.Is(err, usecase.ErrInvalidTime) {
		t.Fatalf("expected ErrInvalidTime, got %v", err)
	}
}

func TestBookingUsecase_CreateBooking_DSTGap(t *testing.T) {
	// Владелец в America/New_York: 2026-03-08 02:30 — несуществующее время (gap).
	r := fake.NewRepositories()
	ny := fake.NewNYOwner("6ba7b810-9dad-41d4-a716-446655443900", "Nina", "nina@example.com")
	ny.WorkingDays = fake.AllWorkingDays()
	r.AddOwner(ny)
	ev := r.AddEvent(&domain.Event{UUID: "6ba7b810-9dad-41d4-a716-446655443901", OwnerID: ny.ID, Name: "Вебинар", DurationMinutes: 30, IsActive: true})

	clock := fake.NewFixedClock(fake.NY(2026, time.March, 8, 0, 0))
	uc := newBookingUsecase(r, clock)

	_, err := uc.CreateBooking(testCtx, bookingInput(ny.UUID, ev.UUID, "Alice", "alice@example.com", "2026-03-08", "02:30"))
	if !errors.Is(err, usecase.ErrAmbiguousTime) {
		t.Fatalf("expected ErrAmbiguousTime, got %v", err)
	}
	if r.Bookings.Count() != 0 {
		t.Fatal("no booking must be created")
	}
}

func TestBookingUsecase_CreateBooking_DSTFold(t *testing.T) {
	// 2026-11-01 01:30 в America/New_York — неоднозначное время (fold).
	r := fake.NewRepositories()
	ny := fake.NewNYOwner("6ba7b810-9dad-41d4-a716-446655443900", "Nina", "nina@example.com")
	ny.WorkingDays = fake.AllWorkingDays()
	r.AddOwner(ny)
	ev := r.AddEvent(&domain.Event{UUID: "6ba7b810-9dad-41d4-a716-446655443901", OwnerID: ny.ID, Name: "Вебинар", DurationMinutes: 30, IsActive: true})

	clock := fake.NewFixedClock(fake.NY(2026, time.November, 1, 0, 0))
	uc := newBookingUsecase(r, clock)

	_, err := uc.CreateBooking(testCtx, bookingInput(ny.UUID, ev.UUID, "Alice", "alice@example.com", "2026-11-01", "01:30"))
	if !errors.Is(err, usecase.ErrAmbiguousTime) {
		t.Fatalf("expected ErrAmbiguousTime, got %v", err)
	}
	if r.Bookings.Count() != 0 {
		t.Fatal("no booking must be created")
	}
}
