package usecase_test

import (
	"errors"
	"testing"
	"time"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/usecase"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/usecase/fake"
)

// Дни в окне 2026-08-10…2026-08-23: будни 10-14 и 17-21, выходные 15-16, 22-23.
func TestSlotUsecase_GenerateSlots_14DayWindow(t *testing.T) {
	r, bob, _ := fake.SeedBob()
	uc := newSlotUsecase(r, fake.NewFixedClock(fake.Today))

	slots, err := uc.GenerateSlots(testCtx, bob.UUID, fake.StandardEventUUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(slots) != 14 {
		t.Fatalf("expected 14 days, got %d", len(slots))
	}

	weekend := map[string]bool{
		"2026-08-15": true, "2026-08-16": true,
		"2026-08-22": true, "2026-08-23": true,
	}
	for i := range 14 {
		date := fake.Today.AddDate(0, 0, i).Format("2006-01-02")
		day, ok := slots[date]
		if !ok {
			t.Fatalf("missing day %s", date)
		}
		if weekend[date] {
			if len(day) != 0 {
				t.Fatalf("weekend %s must have no slots, got %d", date, len(day))
			}
			continue
		}
		if len(day) != 18 {
			t.Fatalf("day %s expected 18 slots, got %d", date, len(day))
		}
		if day[0].StartTime.Format("15:04") != "09:00" || day[17].StartTime.Format("15:04") != "17:30" {
			t.Fatalf("day %s unexpected boundaries: %s..%s", date, day[0].StartTime.Format("15:04"), day[17].StartTime.Format("15:04"))
		}
		for _, s := range day {
			if s.Status != domain.SlotStatusAvailable {
				t.Fatalf("day %s slot %s must be available, got %+v", date, s.StartTime.Format("15:04"), s)
			}
		}
	}
}

func TestSlotUsecase_GenerateSlots_15MinuteEvent(t *testing.T) {
	r, bob, _ := fake.SeedBob()
	uc := newSlotUsecase(r, fake.NewFixedClock(fake.Today))

	slots, err := uc.GenerateSlots(testCtx, bob.UUID, fake.ShortEventUUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	day := slots["2026-08-12"]
	if len(day) != 36 {
		t.Fatalf("expected 36 slots, got %d", len(day))
	}
	if day[0].StartTime.Format("15:04") != "09:00" || day[35].StartTime.Format("15:04") != "17:45" || day[35].EndTime.Format("15:04") != "18:00" {
		t.Fatalf("unexpected boundaries: %s..%s (end %s)", day[0].StartTime.Format("15:04"), day[35].StartTime.Format("15:04"), day[35].EndTime.Format("15:04"))
	}
}

func TestSlotUsecase_GenerateSlots_45MinuteEvent(t *testing.T) {
	r, bob, _ := fake.SeedBob()
	added := r.AddEvent(&domain.Event{UUID: "6ba7b810-9dad-41d4-a716-446655443500", OwnerID: bob.ID, Name: "Консультация", DurationMinutes: 45, IsActive: true})
	uc := newSlotUsecase(r, fake.NewFixedClock(fake.Today))

	slots, err := uc.GenerateSlots(testCtx, bob.UUID, added.UUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	day := slots["2026-08-12"]
	if len(day) != 12 {
		t.Fatalf("expected 12 slots, got %d", len(day))
	}
	last := day[len(day)-1]
	if day[0].StartTime.Format("15:04") != "09:00" || last.StartTime.Format("15:04") != "17:15" || last.EndTime.Format("15:04") != "18:00" {
		t.Fatalf("unexpected boundaries: %s..%s (end %s)", day[0].StartTime.Format("15:04"), last.StartTime.Format("15:04"), last.EndTime.Format("15:04"))
	}
	for i := 1; i < len(day); i++ {
		if day[i].StartTime.Sub(day[i-1].StartTime) != 45*time.Minute {
			t.Fatalf("expected 45m step between slots %d and %d", i-1, i)
		}
	}
}

func TestSlotUsecase_GenerateSlots_FirstAndLastDay(t *testing.T) {
	r, bob, _ := fake.SeedBob()
	uc := newSlotUsecase(r, fake.NewFixedClock(fake.Today))

	slots, err := uc.GenerateSlots(testCtx, bob.UUID, fake.StandardEventUUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := slots["2026-08-10"]; !ok {
		t.Fatal("missing first day 2026-08-10")
	}
	if _, ok := slots["2026-08-23"]; !ok {
		t.Fatal("missing last day 2026-08-23")
	}
	if len(slots["2026-08-23"]) != 0 {
		t.Fatalf("2026-08-23 is Sunday, must be empty")
	}
	if len(slots["2026-08-21"]) != 18 {
		t.Fatalf("expected 18 slots on Friday 2026-08-21, got %d", len(slots["2026-08-21"]))
	}
}

func TestSlotUsecase_GenerateSlots_PastSlotsUnavailable(t *testing.T) {
	r, bob, _ := fake.SeedBob()
	clock := fake.NewFixedClock(fake.Moscow(2026, time.August, 10, 10, 15))
	uc := newSlotUsecase(r, clock)

	slots, err := uc.GenerateSlots(testCtx, bob.UUID, fake.StandardEventUUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	day := slots["2026-08-10"]
	for _, hhmm := range []string{"09:00", "09:30", "10:00"} {
		if s := slotAt(day, hhmm); s.Status != domain.SlotStatusUnavailable || s.Reason != domain.UnavailableReasonPast {
			t.Fatalf("slot %s expected past/unavailable, got %+v", hhmm, s)
		}
	}
	if s := slotAt(day, "10:30"); s.Status != domain.SlotStatusAvailable {
		t.Fatalf("slot 10:30 expected available, got %+v", s)
	}
	// Все слоты следующего дня доступны.
	for _, s := range slots["2026-08-11"] {
		if s.Status != domain.SlotStatusAvailable {
			t.Fatalf("slot %s on 2026-08-11 must be available, got %+v", s.StartTime.Format("15:04"), s)
		}
	}
}

func TestSlotUsecase_GenerateSlots_BookedSlotUnavailable(t *testing.T) {
	r, bob, events := fake.SeedBob()
	std := eventByUUID(events, fake.StandardEventUUID)
	guest := r.AddGuest("Alice", "alice@example.com")
	r.AddBooking(bob.ID, guest.ID, std.ID, fake.Moscow(2026, time.August, 12, 10, 0), fake.Moscow(2026, time.August, 12, 10, 30))
	uc := newSlotUsecase(r, fake.NewFixedClock(fake.Today))

	slots, err := uc.GenerateSlots(testCtx, bob.UUID, fake.StandardEventUUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	day := slots["2026-08-12"]
	if s := slotAt(day, "10:00"); s.Status != domain.SlotStatusUnavailable || s.Reason != domain.UnavailableReasonBooked {
		t.Fatalf("slot 10:00 expected booked/unavailable, got %+v", s)
	}
	for _, hhmm := range []string{"09:30", "10:30"} {
		if s := slotAt(day, hhmm); s.Status != domain.SlotStatusAvailable {
			t.Fatalf("slot %s expected available, got %+v", hhmm, s)
		}
	}
}

func TestSlotUsecase_GenerateSlots_30MinBookingBlocksTwo15MinSlots(t *testing.T) {
	r, bob, events := fake.SeedBob()
	std := eventByUUID(events, fake.StandardEventUUID)
	guest := r.AddGuest("Alice", "alice@example.com")
	r.AddBooking(bob.ID, guest.ID, std.ID, fake.Moscow(2026, time.August, 12, 10, 0), fake.Moscow(2026, time.August, 12, 10, 30))
	uc := newSlotUsecase(r, fake.NewFixedClock(fake.Today))

	slots, err := uc.GenerateSlots(testCtx, bob.UUID, fake.ShortEventUUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	day := slots["2026-08-12"]
	for _, hhmm := range []string{"10:00", "10:15"} {
		if s := slotAt(day, hhmm); s.Status != domain.SlotStatusUnavailable || s.Reason != domain.UnavailableReasonBooked {
			t.Fatalf("slot %s expected booked/unavailable, got %+v", hhmm, s)
		}
	}
	for _, hhmm := range []string{"09:45", "10:30"} {
		if s := slotAt(day, hhmm); s.Status != domain.SlotStatusAvailable {
			t.Fatalf("slot %s expected available, got %+v", hhmm, s)
		}
	}
}

func TestSlotUsecase_GenerateSlots_AnotherTimezone(t *testing.T) {
	r, _, _ := fake.SeedBob()
	ny := fake.NewNYOwner("6ba7b810-9dad-41d4-a716-446655443600", "Nina", "nina@example.com")
	r.AddOwner(ny)
	ev := r.AddEvent(&domain.Event{UUID: "6ba7b810-9dad-41d4-a716-446655443601", OwnerID: ny.ID, Name: "Вебинар", DurationMinutes: 30, IsActive: true})

	// now = понедельник 2026-08-10 12:00 в Нью-Йорке (не в начале дня).
	clock := fake.NewFixedClock(fake.NY(2026, time.August, 10, 12, 0))
	uc := newSlotUsecase(r, clock)

	slots, err := uc.GenerateSlots(testCtx, ny.UUID, ev.UUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(slots) != 14 {
		t.Fatalf("expected 14 days, got %d", len(slots))
	}
	day := slots["2026-08-10"]
	if len(day) != 16 {
		t.Fatalf("expected 16 slots (09:00..16:30), got %d", len(day))
	}
	if day[0].StartTime.Format("15:04") != "09:00" || day[15].StartTime.Format("15:04") != "16:30" {
		t.Fatalf("unexpected boundaries: %s..%s", day[0].StartTime.Format("15:04"), day[15].StartTime.Format("15:04"))
	}
	if loc := day[0].StartTime.Location().String(); loc != "America/New_York" {
		t.Fatalf("expected local timezone America/New_York, got %s", loc)
	}
	// Выходные в Нью-Йорке.
	if len(slots["2026-08-15"]) != 0 || len(slots["2026-08-16"]) != 0 {
		t.Fatal("NY weekend must be empty")
	}
}

func TestSlotUsecase_GenerateSlots_EventNotOwned(t *testing.T) {
	r, bob, _ := fake.SeedBob()
	alice := fake.NewMoscowOwner("6ba7b810-9dad-41d4-a716-446655443700", "Alice Smith", "alice@example.com")
	r.AddOwner(alice)
	aliceEvent := r.AddEvent(&domain.Event{UUID: "6ba7b810-9dad-41d4-a716-446655443701", OwnerID: alice.ID, Name: "Её встреча", DurationMinutes: 30, IsActive: true})
	uc := newSlotUsecase(r, fake.NewFixedClock(fake.Today))

	_, err := uc.GenerateSlots(testCtx, bob.UUID, aliceEvent.UUID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSlotUsecase_GenerateSlots_InvalidEventUUID(t *testing.T) {
	r, bob, _ := fake.SeedBob()
	uc := newSlotUsecase(r, fake.NewFixedClock(fake.Today))

	_, err := uc.GenerateSlots(testCtx, bob.UUID, "not-a-uuid")
	if !errors.Is(err, usecase.ErrInvalidEventUUID) {
		t.Fatalf("expected ErrInvalidEventUUID, got %v", err)
	}
}
