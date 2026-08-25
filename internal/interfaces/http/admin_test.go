package http

import (
	"testing"
	"time"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/usecase/fake"
)

func TestAdminGetOwner(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	var o OwnerResponse
	status := s.doJSON("GET", "/api/admin", nil, &o)
	assertStatus(t, status, 200)
	if o.UUID != fake.BobUUID || o.Name != "Bob Jones" {
		t.Fatalf("unexpected admin owner: %+v", o)
	}
}

func TestAdminListUpcomingBookings(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	guest := s.repos.AddGuest("Jane Roe", "jane@example.com")

	past := fake.Moscow(2026, time.August, 8, 10, 0)
	s.repos.AddBooking(1, guest.ID, 2, past, past.Add(30*time.Minute))

	future1 := fake.Moscow(2026, time.August, 12, 10, 0)
	s.repos.AddBooking(1, guest.ID, 1, future1, future1.Add(15*time.Minute))

	future2 := fake.Moscow(2026, time.August, 11, 11, 0)
	s.repos.AddBooking(1, guest.ID, 2, future2, future2.Add(30*time.Minute))

	var out bookingsAdminResponse
	status := s.doJSON("GET", "/api/admin/bookings", nil, &out)
	assertStatus(t, status, 200)
	if len(out.Bookings) != 2 {
		t.Fatalf("bookings count = %d, want 2", len(out.Bookings))
	}
	if out.Bookings[0].StartTime != "2026-08-11T08:00:00Z" || out.Bookings[1].StartTime != "2026-08-12T07:00:00Z" {
		t.Fatalf("bookings not sorted or unexpected: %+v", out.Bookings)
	}
	if out.Bookings[0].GuestEmail != "jane@example.com" {
		t.Fatalf("guest_email missing: %+v", out.Bookings[0])
	}
}

func TestAdminListEventsIncludesInactive(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	inactive := &domain.Event{UUID: "6ba7b810-9dad-41d4-a716-446655440030", OwnerID: 1, Name: "Архивная", DurationMinutes: 30, IsActive: false}
	s.repos.AddEvent(inactive)

	var out eventsResponse
	status := s.doJSON("GET", "/api/admin/events", nil, &out)
	assertStatus(t, status, 200)
	if len(out.Events) != 3 {
		t.Fatalf("events count = %d, want 3", len(out.Events))
	}
}

func TestAdminCreateEvent(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	status, data := s.do("POST", "/api/admin/events", map[string]any{
		"name": "Консультация", "description": "Разбор", "duration_minutes": 45,
	})
	assertStatus(t, status, 201)
	var created EventResponse
	if err := jsonUnmarshal(data, &created); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !isValidUUID(created.UUID) || created.Name != "Консультация" || created.DurationMinutes != 45 || !created.IsActive {
		t.Fatalf("unexpected created event: %+v", created)
	}
}

func TestAdminCreateEventDuplicate(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	status, data := s.do("POST", "/api/admin/events", map[string]any{
		"name": "Короткая встреча", "duration_minutes": 15,
	})
	assertError(t, status, data, 409, "conflict", "event name already exists")
}

func TestAdminCreateEventInvalidDuration(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	status, data := s.do("POST", "/api/admin/events", map[string]any{
		"name": "Консультация", "duration_minutes": -5,
	})
	assertError(t, status, data, 400, "invalid_input", "invalid duration")
}
