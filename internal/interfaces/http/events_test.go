package http

import (
	"testing"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/usecase/fake"
)

func TestListEvents(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))

	var out eventsResponse
	status := s.doJSON("GET", "/api/owners/"+fake.BobUUID+"/events", nil, &out)
	assertStatus(t, status, 200)
	if len(out.Events) != 2 {
		t.Fatalf("events count = %d, want 2", len(out.Events))
	}
	byName := map[string]EventResponse{}
	for _, e := range out.Events {
		byName[e.Name] = e
	}
	short := byName["Короткая встреча"]
	if short.UUID != fake.ShortEventUUID || short.DurationMinutes != 15 || !short.IsActive {
		t.Fatalf("unexpected short event: %+v", short)
	}
	std := byName["Стандартная встреча"]
	if std.UUID != fake.StandardEventUUID || std.DurationMinutes != 30 || !std.IsActive {
		t.Fatalf("unexpected standard event: %+v", std)
	}
}

func TestListEventsHidesInactive(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	hidden := fake.NewMoscowOwner("6ba7b810-9dad-41d4-a716-446655440010", "Alice", "alice@example.com")
	s.repos.AddOwner(hidden)

	var out eventsResponse
	status := s.doJSON("GET", "/api/owners/"+hidden.UUID+"/events", nil, &out)
	assertStatus(t, status, 200)
	if len(out.Events) != 0 {
		t.Fatalf("events count = %d, want 0", len(out.Events))
	}
}

func TestListEventsOwnerNotFound(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	status, data := s.do("GET", "/api/owners/00000000-0000-4000-8000-00000000dead/events", nil)
	assertError(t, status, data, 404, "not_found", "owner not found")
}

func TestCreateEvent(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))

	status, data := s.do("POST", "/api/owners/"+fake.BobUUID+"/events", map[string]any{
		"name": "Консультация", "description": "Разбор кода", "duration_minutes": 45,
	})
	assertStatus(t, status, 201)
	var created EventResponse
	if err := jsonUnmarshal(data, &created); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !isValidUUID(created.UUID) {
		t.Fatalf("uuid %q is not valid v4", created.UUID)
	}
	if created.Name != "Консультация" || created.Description != "Разбор кода" ||
		created.DurationMinutes != 45 || !created.IsActive {
		t.Fatalf("unexpected created event: %+v", created)
	}
}

func TestCreateEventDuplicate(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	status, data := s.do("POST", "/api/owners/"+fake.BobUUID+"/events", map[string]any{
		"name": "Короткая встреча", "duration_minutes": 15,
	})
	assertError(t, status, data, 409, "conflict", "event name already exists")
}

func TestCreateEventSameNameDifferentOwner(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	alice := fake.NewMoscowOwner("6ba7b810-9dad-41d4-a716-446655440011", "Alice Smith", "alice@example.com")
	s.repos.AddOwner(alice)

	status, _ := s.do("POST", "/api/owners/"+alice.UUID+"/events", map[string]any{
		"name": "Стандартная встреча", "duration_minutes": 30,
	})
	assertStatus(t, status, 201)
}

func TestCreateEventEmptyName(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	status, data := s.do("POST", "/api/owners/"+fake.BobUUID+"/events", map[string]any{
		"name": "  ", "duration_minutes": 30,
	})
	assertError(t, status, data, 400, "invalid_input", "name is required")
}

func TestCreateEventInvalidDuration(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	status, data := s.do("POST", "/api/owners/"+fake.BobUUID+"/events", map[string]any{
		"name": "Консультация", "duration_minutes": 0,
	})
	assertError(t, status, data, 400, "invalid_input", "invalid duration")
}

func TestCreateEventOwnerNotFound(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	status, data := s.do("POST", "/api/owners/00000000-0000-4000-8000-00000000dead/events", map[string]any{
		"name": "Консультация", "duration_minutes": 45,
	})
	assertError(t, status, data, 404, "not_found", "owner not found")
}
