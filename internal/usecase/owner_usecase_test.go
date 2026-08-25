package usecase_test

import (
	"errors"
	"testing"
	"time"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/usecase"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/usecase/fake"
)

func TestOwnerUsecase_ListActiveOwners(t *testing.T) {
	r, bob, _ := fake.SeedBob()
	alice := fake.NewMoscowOwner("6ba7b810-9dad-41d4-a716-446655443100", "Alice Smith", "alice@example.com")
	r.AddOwner(alice)

	uc := newOwnerUsecase(r)
	owners, err := uc.ListActiveOwners(testCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(owners) != 2 {
		t.Fatalf("expected 2 owners, got %d", len(owners))
	}
	names := map[string]bool{}
	for _, o := range owners {
		names[o.Name] = true
	}
	if !names["Bob Jones"] || !names["Alice Smith"] {
		t.Fatalf("expected Bob and Alice, got %v", names)
	}
	_ = bob
}

func TestOwnerUsecase_ListActiveOwners_Empty(t *testing.T) {
	r := fake.NewRepositories()
	uc := newOwnerUsecase(r)
	owners, err := uc.ListActiveOwners(testCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if owners == nil || len(owners) != 0 {
		t.Fatalf("expected empty non-nil list, got %#v", owners)
	}
}

func TestOwnerUsecase_ListActiveOwners_SkipsInactive(t *testing.T) {
	r, _, _ := fake.SeedBob()
	inactive := fake.NewMoscowOwner("6ba7b810-9dad-41d4-a716-446655443200", "Carol", "carol@example.com")
	inactive.IsActive = false
	r.AddOwner(inactive)

	uc := newOwnerUsecase(r)
	owners, err := uc.ListActiveOwners(testCtx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(owners) != 1 || owners[0].Name != "Bob Jones" {
		t.Fatalf("expected only Bob, got %#v", owners)
	}
}

func TestOwnerUsecase_GetByUUID_HappyPath(t *testing.T) {
	r, bob, _ := fake.SeedBob()
	uc := newOwnerUsecase(r)
	got, err := uc.GetByUUID(testCtx, bob.UUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "Bob Jones" || got.Email != "bob@example.com" {
		t.Fatalf("unexpected owner: %#v", got)
	}
}

func TestOwnerUsecase_GetByUUID_NotFound(t *testing.T) {
	r, _, _ := fake.SeedBob()
	uc := newOwnerUsecase(r)
	_, err := uc.GetByUUID(testCtx, "00000000-0000-4000-8000-000000000000")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestOwnerUsecase_GetByUUID_InvalidUUID(t *testing.T) {
	r, _, _ := fake.SeedBob()
	uc := newOwnerUsecase(r)
	_, err := uc.GetByUUID(testCtx, "not-a-uuid")
	if !errors.Is(err, usecase.ErrInvalidOwnerUUID) {
		t.Fatalf("expected ErrInvalidOwnerUUID, got %v", err)
	}
}

func TestOwnerUsecase_CreateOwner_HappyPath(t *testing.T) {
	r, _, _ := fake.SeedBob()
	uc := newOwnerUsecase(r)

	owner, err := uc.CreateOwner(testCtx, "Alice Smith", "alice@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if owner.Name != "Alice Smith" {
		t.Fatalf("unexpected name: %s", owner.Name)
	}
	if !isUUIDv4(owner.UUID) {
		t.Fatalf("expected valid UUID v4, got %s", owner.UUID)
	}
	if owner.WorkStart != "09:00" || owner.WorkEnd != "18:00" || owner.Timezone != "Europe/Moscow" {
		t.Fatalf("unexpected default settings: %+v", owner)
	}
	if !owner.WorkingDays["mon"] || owner.WorkingDays["sun"] {
		t.Fatalf("unexpected working days: %v", owner.WorkingDays)
	}

	events, err := r.Events.GetByOwnerID(testCtx, owner.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 default events, got %d", len(events))
	}
	byName := map[string]int{}
	for _, e := range events {
		byName[e.Name] = e.DurationMinutes
	}
	if byName["Короткая встреча"] != 15 || byName["Стандартная встреча"] != 30 {
		t.Fatalf("unexpected default events: %v", byName)
	}
}

func TestOwnerUsecase_CreateOwner_InvalidEmail(t *testing.T) {
	r := fake.NewRepositories()
	uc := newOwnerUsecase(r)

	_, err := uc.CreateOwner(testCtx, "Alice", "not-an-email")
	if !errors.Is(err, usecase.ErrInvalidEmail) {
		t.Fatalf("expected ErrInvalidEmail, got %v", err)
	}
	if owners, _ := r.Owners.GetAll(testCtx); len(owners) != 0 {
		t.Fatalf("owner must not be created")
	}
}

func TestOwnerUsecase_CreateOwner_EmptyName(t *testing.T) {
	r := fake.NewRepositories()
	uc := newOwnerUsecase(r)

	_, err := uc.CreateOwner(testCtx, "", "alice@example.com")
	if !errors.Is(err, usecase.ErrNameRequired) {
		t.Fatalf("expected ErrNameRequired, got %v", err)
	}
	if owners, _ := r.Owners.GetAll(testCtx); len(owners) != 0 {
		t.Fatalf("owner must not be created")
	}
}

func TestOwnerUsecase_CreateOwner_EmailExists(t *testing.T) {
	r, _, _ := fake.SeedBob()
	uc := newOwnerUsecase(r)

	_, err := uc.CreateOwner(testCtx, "Bob Clone", "bob@example.com")
	if !errors.Is(err, usecase.ErrEmailExists) {
		t.Fatalf("expected ErrEmailExists, got %v", err)
	}
	if owners, _ := r.Owners.GetAll(testCtx); len(owners) != 1 {
		t.Fatalf("owner must not be created")
	}
}

func TestOwnerUsecase_ListBookings_SortedByStartTime(t *testing.T) {
	r, bob, events := fake.SeedBob()
	std := eventByUUID(events, fake.StandardEventUUID)
	guest := r.AddGuest("Alice", "alice@example.com")
	r.AddBooking(bob.ID, guest.ID, std.ID, fake.Moscow(2026, time.August, 12, 10, 0), fake.Moscow(2026, time.August, 12, 10, 30))
	r.AddBooking(bob.ID, guest.ID, std.ID, fake.Moscow(2026, time.August, 11, 9, 0), fake.Moscow(2026, time.August, 11, 9, 30))

	uc := newOwnerUsecase(r)
	bookings, err := uc.ListBookings(testCtx, bob.UUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bookings) != 2 {
		t.Fatalf("expected 2 bookings, got %d", len(bookings))
	}
	if bookings[0].StartTime.Format("01-02") != "08-11" || bookings[1].StartTime.Format("01-02") != "08-12" {
		t.Fatalf("expected sorted by start_time, got %s / %s", bookings[0].StartTime.Format("01-02"), bookings[1].StartTime.Format("01-02"))
	}
	if bookings[0].GuestName != "Alice" || bookings[0].EventName != "Стандартная встреча" {
		t.Fatalf("unexpected booking fields: %+v", bookings[0])
	}
}

func TestOwnerUsecase_ListBookings_Empty(t *testing.T) {
	r, bob, _ := fake.SeedBob()
	uc := newOwnerUsecase(r)
	bookings, err := uc.ListBookings(testCtx, bob.UUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bookings == nil || len(bookings) != 0 {
		t.Fatalf("expected empty list, got %#v", bookings)
	}
}
