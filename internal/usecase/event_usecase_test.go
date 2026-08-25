package usecase_test

import (
	"errors"
	"testing"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/usecase"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/usecase/fake"
)

func TestEventUsecase_ListActiveByOwner(t *testing.T) {
	r, bob, _ := fake.SeedBob()
	uc := newEventUsecase(r)

	got, err := uc.ListActiveByOwner(testCtx, bob.UUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 events, got %d", len(got))
	}
	byName := map[string]domain.Event{}
	for _, e := range got {
		byName[e.Name] = e
	}
	if e, ok := byName["Короткая встреча"]; !ok || e.DurationMinutes != 15 {
		t.Fatalf("missing short event: %#v", got)
	}
	if e, ok := byName["Стандартная встреча"]; !ok || e.DurationMinutes != 30 {
		t.Fatalf("missing standard event: %#v", got)
	}
}

func TestEventUsecase_ListActiveByOwner_HidesInactive(t *testing.T) {
	r, bob, _ := fake.SeedBob()
	r.AddEvent(&domain.Event{UUID: "6ba7b810-9dad-41d4-a716-446655443300", OwnerID: bob.ID, Name: "Консультация", DurationMinutes: 45, IsActive: false})

	uc := newEventUsecase(r)
	got, err := uc.ListActiveByOwner(testCtx, bob.UUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected only 2 active events, got %d", len(got))
	}
}

func TestEventUsecase_ListActiveByOwner_OwnerNotFound(t *testing.T) {
	r, _, _ := fake.SeedBob()
	uc := newEventUsecase(r)

	_, err := uc.ListActiveByOwner(testCtx, "00000000-0000-4000-8000-000000000000")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestEventUsecase_CreateForOwner_HappyPath(t *testing.T) {
	r, bob, _ := fake.SeedBob()
	uc := newEventUsecase(r)

	ev, err := uc.CreateForOwner(testCtx, bob.UUID, "Консультация", "45 минут", 45)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isUUIDv4(ev.UUID) {
		t.Fatalf("expected valid UUID v4, got %s", ev.UUID)
	}
	if ev.Name != "Консультация" || ev.DurationMinutes != 45 || !ev.IsActive {
		t.Fatalf("unexpected event: %+v", ev)
	}
	if ev.OwnerID != bob.ID {
		t.Fatalf("wrong owner: %d != %d", ev.OwnerID, bob.ID)
	}
}

func TestEventUsecase_CreateForOwner_DuplicateName(t *testing.T) {
	r, bob, _ := fake.SeedBob()
	uc := newEventUsecase(r)

	_, err := uc.CreateForOwner(testCtx, bob.UUID, "Короткая встреча", "", 20)
	if !errors.Is(err, usecase.ErrEventNameExists) {
		t.Fatalf("expected ErrEventNameExists, got %v", err)
	}
}

func TestEventUsecase_CreateForOwner_SameNameDifferentOwner(t *testing.T) {
	r, _, _ := fake.SeedBob()
	alice := fake.NewMoscowOwner("6ba7b810-9dad-41d4-a716-446655443400", "Alice Smith", "alice@example.com")
	r.AddOwner(alice)
	uc := newEventUsecase(r)

	ev, err := uc.CreateForOwner(testCtx, alice.UUID, "Короткая встреча", "", 15)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if ev.OwnerID != alice.ID {
		t.Fatalf("wrong owner: %d != %d", ev.OwnerID, alice.ID)
	}
}

func TestEventUsecase_CreateForOwner_InvalidDuration(t *testing.T) {
	r, bob, _ := fake.SeedBob()
	uc := newEventUsecase(r)

	_, err := uc.CreateForOwner(testCtx, bob.UUID, "Событие", "", 0)
	if !errors.Is(err, usecase.ErrInvalidDuration) {
		t.Fatalf("expected ErrInvalidDuration, got %v", err)
	}
}

func TestEventUsecase_CreateForOwner_EmptyName(t *testing.T) {
	r, bob, _ := fake.SeedBob()
	uc := newEventUsecase(r)

	_, err := uc.CreateForOwner(testCtx, bob.UUID, "", "", 30)
	if !errors.Is(err, usecase.ErrNameRequired) {
		t.Fatalf("expected ErrNameRequired, got %v", err)
	}
}

func TestEventUsecase_CreateForOwner_OwnerNotFound(t *testing.T) {
	r, _, _ := fake.SeedBob()
	uc := newEventUsecase(r)

	_, err := uc.CreateForOwner(testCtx, "00000000-0000-4000-8000-000000000000", "Событие", "", 30)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
