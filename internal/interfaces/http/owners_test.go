package http

import (
	"testing"
	"time"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/usecase/fake"
)

func TestHealthz(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	status, data := s.do("GET", "/healthz", nil)
	assertStatus(t, status, 200)
	var out struct {
		Status string `json:"status"`
	}
	if err := jsonUnmarshal(data, &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Status != "ok" {
		t.Fatalf("status = %q, want ok", out.Status)
	}
}

func TestListOwnersEmpty(t *testing.T) {
	s := newEmptyTestServer(t, fake.NewFixedClock(fake.Today))
	status, data := s.do("GET", "/api/owners", nil)
	assertStatus(t, status, 200)
	if string(data) != "[]\n" && string(data) != "[]" {
		t.Fatalf("body = %s, want empty array", data)
	}
}

func TestListOwners(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))

	var owners []OwnerSummary
	status := s.doJSON("GET", "/api/owners", nil, &owners)
	assertStatus(t, status, 200)
	if len(owners) != 1 {
		t.Fatalf("owners count = %d, want 1", len(owners))
	}
	if owners[0].UUID != fake.BobUUID || owners[0].Name != "Bob Jones" {
		t.Fatalf("unexpected owner: %+v", owners[0])
	}
}

func TestListOwnersSkipsInactive(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	inactive := fake.NewMoscowOwner("6ba7b810-9dad-41d4-a716-446655440099", "Carol", "carol@example.com")
	inactive.IsActive = false
	s.repos.AddOwner(inactive)

	var owners []OwnerSummary
	status := s.doJSON("GET", "/api/owners", nil, &owners)
	assertStatus(t, status, 200)
	if len(owners) != 1 || owners[0].UUID != fake.BobUUID {
		t.Fatalf("inactive owner should be hidden, got %+v", owners)
	}
}

func TestGetOwner(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))

	var o OwnerResponse
	status := s.doJSON("GET", "/api/owners/"+fake.BobUUID, nil, &o)
	assertStatus(t, status, 200)
	if o.UUID != fake.BobUUID || o.Name != "Bob Jones" {
		t.Fatalf("unexpected owner: %+v", o)
	}
	if o.Settings.WorkStart != "09:00" || o.Settings.WorkEnd != "18:00" || o.Settings.Timezone != "Europe/Moscow" {
		t.Fatalf("unexpected settings: %+v", o.Settings)
	}
	want := []string{"mon", "tue", "wed", "thu", "fri"}
	if !equalStrings(o.Settings.WorkingDays, want) {
		t.Fatalf("working_days = %v, want %v", o.Settings.WorkingDays, want)
	}
}

func TestGetOwnerNotFound(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	status, data := s.do("GET", "/api/owners/00000000-0000-4000-8000-00000000dead", nil)
	assertError(t, status, data, 404, "not_found", "owner not found")
}

func TestGetOwnerInvalidUUID(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	status, data := s.do("GET", "/api/owners/not-a-uuid", nil)
	assertError(t, status, data, 400, "invalid_input", "invalid owner uuid")
}

func TestCreateOwner(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))

	status, data := s.do("POST", "/api/owners", map[string]any{
		"name": "Alice Smith", "email": "alice@example.com",
	})
	assertStatus(t, status, 201)
	var created OwnerResponse
	if err := jsonUnmarshal(data, &created); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if created.UUID == "" {
		t.Fatal("empty uuid in create response")
	}
	if created.Name != "Alice Smith" {
		t.Fatalf("name = %q", created.Name)
	}
	if !isValidUUID(created.UUID) {
		t.Fatalf("uuid %q is not valid v4", created.UUID)
	}
	if created.Settings.WorkStart != "09:00" || created.Settings.Timezone != "Europe/Moscow" {
		t.Fatalf("default settings not applied: %+v", created.Settings)
	}

	var eventsOut eventsResponse
	status = s.doJSON("GET", "/api/owners/"+created.UUID+"/events", nil, &eventsOut)
	assertStatus(t, status, 200)
	if len(eventsOut.Events) != 2 {
		t.Fatalf("default events count = %d, want 2", len(eventsOut.Events))
	}
}

func TestCreateOwnerDuplicateEmail(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	status, data := s.do("POST", "/api/owners", map[string]any{
		"name": "Another Bob", "email": "bob@example.com",
	})
	assertError(t, status, data, 409, "conflict", "email already exists")
}

func TestCreateOwnerInvalidEmail(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	status, data := s.do("POST", "/api/owners", map[string]any{
		"name": "Alice", "email": "not-an-email",
	})
	assertError(t, status, data, 400, "invalid_input", "invalid email")
}

func TestCreateOwnerEmptyName(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	status, data := s.do("POST", "/api/owners", map[string]any{
		"name": "  ", "email": "alice@example.com",
	})
	assertError(t, status, data, 400, "invalid_input", "name is required")
}

func TestListOwnerBookingsEmpty(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	status, data := s.do("GET", "/api/owners/"+fake.BobUUID+"/bookings", nil)
	assertStatus(t, status, 200)
	if string(data) != "{\"bookings\":[]}\n" && string(data) != "{\"bookings\":[]}" {
		t.Fatalf("body = %s, want empty bookings array", data)
	}
}

func TestListOwnerBookingsSorted(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	guest := s.repos.AddGuest("Jane Roe", "jane@example.com")
	// Добавляем в обратном порядке: сначала более поздняя, потом более ранняя.
	s.repos.AddBooking(1, guest.ID, 2, fake.Moscow(2026, time.August, 12, 10, 0), fake.Moscow(2026, time.August, 12, 10, 30))
	s.repos.AddBooking(1, guest.ID, 2, fake.Moscow(2026, time.August, 11, 9, 0), fake.Moscow(2026, time.August, 11, 9, 30))

	var out bookingsResponse
	status := s.doJSON("GET", "/api/owners/"+fake.BobUUID+"/bookings", nil, &out)
	assertStatus(t, status, 200)
	if len(out.Bookings) != 2 {
		t.Fatalf("bookings count = %d, want 2", len(out.Bookings))
	}
	if out.Bookings[0].StartTime != "2026-08-11T06:00:00Z" || out.Bookings[1].StartTime != "2026-08-12T07:00:00Z" {
		t.Fatalf("bookings not sorted by start_time: %s, %s", out.Bookings[0].StartTime, out.Bookings[1].StartTime)
	}
}

func TestListOwnerBookings(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	guest := s.repos.AddGuest("Jane Roe", "jane@example.com")
	start := fake.Moscow(2026, time.August, 12, 10, 0)
	s.repos.AddBooking(1, guest.ID, 2, start, start.Add(30*time.Minute))

	var out bookingsResponse
	status := s.doJSON("GET", "/api/owners/"+fake.BobUUID+"/bookings", nil, &out)
	assertStatus(t, status, 200)
	if len(out.Bookings) != 1 {
		t.Fatalf("bookings count = %d, want 1", len(out.Bookings))
	}
	b := out.Bookings[0]
	if b.GuestName != "Jane Roe" || b.EventName != "Стандартная встреча" || b.DurationMinutes != 30 {
		t.Fatalf("unexpected booking: %+v", b)
	}
	if b.StartTime != "2026-08-12T07:00:00Z" || b.EndTime != "2026-08-12T07:30:00Z" {
		t.Fatalf("unexpected times: %s — %s", b.StartTime, b.EndTime)
	}
}

func TestListOwnerBookingsOwnerNotFound(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	status, data := s.do("GET", "/api/owners/00000000-0000-4000-8000-00000000dead/bookings", nil)
	assertError(t, status, data, 404, "not_found", "owner not found")
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
