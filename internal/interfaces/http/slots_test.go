package http

import (
	"testing"
	"time"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/usecase/fake"
)

func TestListSlotsWorkingDay(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))

	var out SlotsResponse
	status := s.doJSON("GET", "/api/owners/"+fake.BobUUID+"/slots?event_uuid="+fake.StandardEventUUID, nil, &out)
	assertStatus(t, status, 200)
	if out.EventUUID != fake.StandardEventUUID || out.EventName != "Стандартная встреча" || out.DurationMinutes != 30 {
		t.Fatalf("unexpected event meta: %+v", out)
	}
	if out.Timezone != "Europe/Moscow" {
		t.Fatalf("timezone = %q", out.Timezone)
	}
	if out.StartDate != "2026-08-10" || out.EndDate != "2026-08-23" {
		t.Fatalf("window = %s..%s", out.StartDate, out.EndDate)
	}

	slots := out.Slots["2026-08-10"]
	if len(slots) != 18 {
		t.Fatalf("2026-08-10 slots count = %d, want 18", len(slots))
	}
	if slots[0].Time != "09:00" || slots[len(slots)-1].Time != "17:30" {
		t.Fatalf("slot bounds = %s..%s", slots[0].Time, slots[len(slots)-1].Time)
	}
	for _, s := range slots {
		if s.Status != "available" {
			t.Fatalf("slot %s status = %q, want available", s.Time, s.Status)
		}
	}

	if weekend := out.Slots["2026-08-15"]; len(weekend) != 0 {
		t.Fatalf("saturday slots = %v, want empty", weekend)
	}
}

func TestSlotsMarkPast(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Moscow(2026, time.August, 10, 10, 15)))

	var out SlotsResponse
	status := s.doJSON("GET", "/api/owners/"+fake.BobUUID+"/slots?event_uuid="+fake.StandardEventUUID, nil, &out)
	assertStatus(t, status, 200)

	slots := out.Slots["2026-08-10"]
	for _, want := range []string{"09:00", "09:30", "10:00"} {
		got := slotByTime(slots, want)
		if got == nil || got.Status != "unavailable" || got.Reason != "past" {
			t.Fatalf("slot %s = %+v, want unavailable/past", want, got)
		}
	}
	next := slotByTime(slots, "10:30")
	if next == nil || next.Status != "available" {
		t.Fatalf("slot 10:30 = %+v, want available", next)
	}
}

func TestSlotsMarkBooked(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	guest := s.repos.AddGuest("Jane Roe", "jane@example.com")
	start := fake.Moscow(2026, time.August, 12, 10, 0)
	s.repos.AddBooking(1, guest.ID, 2, start, start.Add(30*time.Minute))

	var out SlotsResponse
	status := s.doJSON("GET", "/api/owners/"+fake.BobUUID+"/slots?event_uuid="+fake.StandardEventUUID, nil, &out)
	assertStatus(t, status, 200)

	slots := out.Slots["2026-08-12"]
	if got := slotByTime(slots, "10:00"); got == nil || got.Status != "unavailable" || got.Reason != "booked" {
		t.Fatalf("slot 10:00 = %+v, want unavailable/booked", got)
	}
	if got := slotByTime(slots, "10:30"); got == nil || got.Status != "available" {
		t.Fatalf("slot 10:30 = %+v, want available", got)
	}
}

func TestSlotsMarkBookedShortEvent(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	guest := s.repos.AddGuest("Jane Roe", "jane@example.com")
	start := fake.Moscow(2026, time.August, 12, 10, 0)
	s.repos.AddBooking(1, guest.ID, 2, start, start.Add(30*time.Minute))

	var out SlotsResponse
	status := s.doJSON("GET", "/api/owners/"+fake.BobUUID+"/slots?event_uuid="+fake.ShortEventUUID, nil, &out)
	assertStatus(t, status, 200)

	slots := out.Slots["2026-08-12"]
	for _, want := range []string{"10:00", "10:15"} {
		if got := slotByTime(slots, want); got == nil || got.Status != "unavailable" || got.Reason != "booked" {
			t.Fatalf("slot %s = %+v, want unavailable/booked", want, got)
		}
	}
	for _, want := range []string{"09:45", "10:30"} {
		if got := slotByTime(slots, want); got == nil || got.Status != "available" {
			t.Fatalf("slot %s = %+v, want available", want, got)
		}
	}
}

func TestSlotsDifferentTimezone(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.NY(2026, time.August, 10, 0, 0)))
	owner := fake.NewNYOwner("6ba7b810-9dad-41d4-a716-446655440010", "Alice Smith", "alice@example.com")
	s.repos.AddOwner(owner)
	ev := &domain.Event{UUID: "6ba7b810-9dad-41d4-a716-446655440020", OwnerID: owner.ID, Name: "Консультация", DurationMinutes: 30, IsActive: true}
	s.repos.AddEvent(ev)

	var out SlotsResponse
	status := s.doJSON("GET", "/api/owners/"+owner.UUID+"/slots?event_uuid="+ev.UUID, nil, &out)
	assertStatus(t, status, 200)
	if out.Timezone != "America/New_York" {
		t.Fatalf("timezone = %q", out.Timezone)
	}
	slots := out.Slots["2026-08-10"]
	if len(slots) != 16 {
		t.Fatalf("NY slots count = %d, want 16", len(slots))
	}
	if slots[0].Time != "09:00" || slots[len(slots)-1].Time != "16:30" {
		t.Fatalf("NY slot bounds = %s..%s", slots[0].Time, slots[len(slots)-1].Time)
	}
}

func TestSlotsShortEventCount(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))

	var out SlotsResponse
	status := s.doJSON("GET", "/api/owners/"+fake.BobUUID+"/slots?event_uuid="+fake.ShortEventUUID, nil, &out)
	assertStatus(t, status, 200)

	slots := out.Slots["2026-08-10"]
	if len(slots) != 36 {
		t.Fatalf("15-min event slots count = %d, want 36", len(slots))
	}
	if slots[0].Time != "09:00" || slots[len(slots)-1].Time != "17:45" {
		t.Fatalf("slot bounds = %s..%s", slots[0].Time, slots[len(slots)-1].Time)
	}
	if slotByTime(slots, "18:00") != nil {
		t.Fatal("unexpected 18:00 slot for short event")
	}
}

func TestSlotsFortyFiveMinuteEvent(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	ev := &domain.Event{UUID: "6ba7b810-9dad-41d4-a716-446655440040", OwnerID: 1, Name: "Консультация", DurationMinutes: 45, IsActive: true}
	s.repos.AddEvent(ev)

	var out SlotsResponse
	status := s.doJSON("GET", "/api/owners/"+fake.BobUUID+"/slots?event_uuid="+ev.UUID, nil, &out)
	assertStatus(t, status, 200)

	slots := out.Slots["2026-08-10"]
	if slots[0].Time != "09:00" || slots[len(slots)-1].Time != "17:15" {
		t.Fatalf("slot bounds = %s..%s", slots[0].Time, slots[len(slots)-1].Time)
	}
	if slotByTime(slots, "17:30") != nil {
		t.Fatal("unexpected 17:30 slot for 45-min event")
	}
}

func TestSlotsWindowKeys(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))

	var out SlotsResponse
	status := s.doJSON("GET", "/api/owners/"+fake.BobUUID+"/slots?event_uuid="+fake.StandardEventUUID, nil, &out)
	assertStatus(t, status, 200)

	if _, ok := out.Slots["2026-08-10"]; !ok {
		t.Fatal("missing first window key 2026-08-10")
	}
	if _, ok := out.Slots["2026-08-23"]; !ok {
		t.Fatal("missing last window key 2026-08-23")
	}
	if len(out.Slots) != 14 {
		t.Fatalf("window days = %d, want 14", len(out.Slots))
	}
}

func TestSlotsInvalidEventUUID(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	status, data := s.do("GET", "/api/owners/"+fake.BobUUID+"/slots?event_uuid=not-a-uuid", nil)
	assertError(t, status, data, 400, "invalid_input", "invalid event uuid")
}

func TestSlotsEventNotFound(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	status, data := s.do("GET", "/api/owners/"+fake.BobUUID+"/slots?event_uuid=00000000-0000-4000-8000-00000000dead", nil)
	assertError(t, status, data, 404, "not_found", "event not found")
}

func TestSlotsInactiveEvent(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	inactive := &domain.Event{UUID: "6ba7b810-9dad-41d4-a716-446655440030", OwnerID: 1, Name: "Архивная", DurationMinutes: 30, IsActive: false}
	s.repos.AddEvent(inactive)
	status, data := s.do("GET", "/api/owners/"+fake.BobUUID+"/slots?event_uuid="+inactive.UUID, nil)
	assertError(t, status, data, 404, "not_found", "event not found")
}

func TestSlotsOwnerNotFound(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	status, data := s.do("GET", "/api/owners/00000000-0000-4000-8000-00000000dead/slots?event_uuid="+fake.StandardEventUUID, nil)
	assertError(t, status, data, 404, "not_found", "owner not found")
}

func slotByTime(slots []SlotResponse, want string) *SlotResponse {
	for i := range slots {
		if slots[i].Time == want {
			return &slots[i]
		}
	}
	return nil
}
