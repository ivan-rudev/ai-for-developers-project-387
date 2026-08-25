package http

import (
	"testing"
	"time"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/usecase/fake"
)

const (
	ghostUUID = "00000000-0000-4000-8000-00000000dead"
	aliceUUID = "6ba7b810-9dad-41d4-a716-446655440010"
)

func bookingBody(ownerUUID, eventUUID, start string) map[string]any {
	return map[string]any{
		"owner_uuid":  ownerUUID,
		"event_uuid":  eventUUID,
		"guest_name":  "Jane Roe",
		"guest_email": "jane@example.com",
		"date":        "2026-08-12",
		"start_time":  start,
	}
}

func TestCreateBooking(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))

	status, data := s.do("POST", "/api/bookings", bookingBody(fake.BobUUID, fake.StandardEventUUID, "10:00"))
	assertStatus(t, status, 201)
	var out BookingResponse
	if err := jsonUnmarshal(data, &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.OwnerUUID != fake.BobUUID || out.GuestName != "Jane Roe" ||
		out.EventName != "Стандартная встреча" || out.DurationMinutes != 30 {
		t.Fatalf("unexpected booking: %+v", out)
	}
	if out.StartTime != "2026-08-12T07:00:00Z" || out.EndTime != "2026-08-12T07:30:00Z" {
		t.Fatalf("times = %s..%s", out.StartTime, out.EndTime)
	}
	if out.CreatedAt != "2026-08-10T00:00:00Z" {
		t.Fatalf("created_at = %q", out.CreatedAt)
	}
}

func TestCreateBookingAppearsInList(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	status, _ := s.do("POST", "/api/bookings", bookingBody(fake.BobUUID, fake.StandardEventUUID, "10:00"))
	assertStatus(t, status, 201)

	var out bookingsResponse
	status = s.doJSON("GET", "/api/owners/"+fake.BobUUID+"/bookings", nil, &out)
	assertStatus(t, status, 200)
	if len(out.Bookings) != 1 {
		t.Fatalf("bookings count = %d, want 1", len(out.Bookings))
	}
}

func TestCreateBookingReusesGuest(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))

	status, _ := s.do("POST", "/api/bookings", bookingBody(fake.BobUUID, fake.StandardEventUUID, "10:00"))
	assertStatus(t, status, 201)
	status, _ = s.do("POST", "/api/bookings", bookingBody(fake.BobUUID, fake.StandardEventUUID, "11:00"))
	assertStatus(t, status, 201)

	if n := s.repos.Guests.Count(); n != 1 {
		t.Fatalf("guests count = %d, want 1 (guest must be reused)", n)
	}
}

func TestCreateBookingEventNotFound(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	status, data := s.do("POST", "/api/bookings", bookingBody(fake.BobUUID, ghostUUID, "10:00"))
	assertError(t, status, data, 404, "not_found", "event not found")
}

func TestCreateBookingOwnerNotFound(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	status, data := s.do("POST", "/api/bookings", bookingBody(ghostUUID, fake.StandardEventUUID, "10:00"))
	assertError(t, status, data, 404, "not_found", "owner not found")
}

func TestCreateBookingBookedSlot(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	guest := s.repos.AddGuest("Other", "other@example.com")
	start := fake.Moscow(2026, time.August, 12, 10, 0)
	s.repos.AddBooking(1, guest.ID, 2, start, start.Add(30*time.Minute))

	status, data := s.do("POST", "/api/bookings", bookingBody(fake.BobUUID, fake.StandardEventUUID, "10:00"))
	assertStatus(t, status, 409)
	var e struct {
		Error string `json:"error"`
	}
	if err := jsonUnmarshal(data, &e); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if e.Error != "overlap" && e.Error != "slot_unavailable" {
		t.Fatalf("error code = %q, want overlap or slot_unavailable", e.Error)
	}
}

func TestCreateBookingPartialOverlap(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	guest := s.repos.AddGuest("Other", "other@example.com")
	start := fake.Moscow(2026, time.August, 12, 10, 0)
	s.repos.AddBooking(1, guest.ID, 2, start, start.Add(30*time.Minute))

	status, data := s.do("POST", "/api/bookings", bookingBody(fake.BobUUID, fake.ShortEventUUID, "10:15"))
	assertStatus(t, status, 409)
	var e struct {
		Error string `json:"error"`
	}
	if err := jsonUnmarshal(data, &e); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if e.Error != "overlap" && e.Error != "slot_unavailable" {
		t.Fatalf("error code = %q, want overlap or slot_unavailable", e.Error)
	}
}

func TestCreateBookingEndOfWorkingDay(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	status, data := s.do("POST", "/api/bookings", bookingBody(fake.BobUUID, fake.StandardEventUUID, "17:30"))
	assertStatus(t, status, 201)
	var out BookingResponse
	if err := jsonUnmarshal(data, &out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.EndTime != "2026-08-12T15:00:00Z" {
		t.Fatalf("end_time = %q, want 2026-08-12T15:00:00Z", out.EndTime)
	}
}

func TestCreateBookingSundayAndFriday(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))

	body := bookingBody(fake.BobUUID, fake.StandardEventUUID, "10:00")
	body["date"] = "2026-08-23"
	status, data := s.do("POST", "/api/bookings", body)
	assertError(t, status, data, 400, "invalid_input", "selected day is not a working day")

	body["date"] = "2026-08-21"
	status, _ = s.do("POST", "/api/bookings", body)
	assertStatus(t, status, 201)
}

func TestCreateBookingOutsideWindow(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	body := bookingBody(fake.BobUUID, fake.StandardEventUUID, "10:00")
	body["date"] = "2026-08-24"
	status, data := s.do("POST", "/api/bookings", body)
	assertError(t, status, data, 400, "invalid_input", "date is outside available range")
}

func TestCreateBookingSaturday(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	body := bookingBody(fake.BobUUID, fake.StandardEventUUID, "10:00")
	body["date"] = "2026-08-15"
	status, data := s.do("POST", "/api/bookings", body)
	assertError(t, status, data, 400, "invalid_input", "selected day is not a working day")
}

func TestCreateBookingOutsideHours(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	status, data := s.do("POST", "/api/bookings", bookingBody(fake.BobUUID, fake.StandardEventUUID, "08:30"))
	assertError(t, status, data, 400, "invalid_input", "slot is outside working hours")
}

func TestCreateBookingNotAligned(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	status, data := s.do("POST", "/api/bookings", bookingBody(fake.BobUUID, fake.StandardEventUUID, "10:10"))
	assertError(t, status, data, 400, "invalid_input", "slot start is not aligned with event duration")
}

func TestCreateBookingInPast(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Moscow(2026, time.August, 10, 14, 0)))
	body := bookingBody(fake.BobUUID, fake.StandardEventUUID, "13:30")
	body["date"] = "2026-08-10"
	status, data := s.do("POST", "/api/bookings", body)
	assertError(t, status, data, 400, "invalid_input", "slot is in the past")
}

func TestCreateBookingInvalidEmail(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	body := bookingBody(fake.BobUUID, fake.StandardEventUUID, "10:00")
	body["guest_email"] = "not-an-email"
	status, data := s.do("POST", "/api/bookings", body)
	assertError(t, status, data, 400, "invalid_input", "invalid email")
}

func TestCreateBookingWrongOwnerEvent(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	_, aliceEvent := s.addAlice()

	status, data := s.do("POST", "/api/bookings", bookingBody(fake.BobUUID, aliceEvent.UUID, "10:00"))
	assertError(t, status, data, 404, "not_found", "event not found")

	status, data = s.do("POST", "/api/bookings", bookingBody(aliceUUID, fake.StandardEventUUID, "10:00"))
	assertError(t, status, data, 404, "not_found", "event not found")
}

func TestCreateBookingInvalidDateAndTime(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))

	body := bookingBody(fake.BobUUID, fake.StandardEventUUID, "10:00")
	body["date"] = "2026/08/12"
	status, data := s.do("POST", "/api/bookings", body)
	assertError(t, status, data, 400, "invalid_input", "invalid date")

	body = bookingBody(fake.BobUUID, fake.StandardEventUUID, "25:99")
	status, data = s.do("POST", "/api/bookings", body)
	assertError(t, status, data, 400, "invalid_input", "invalid time")
}

func TestCreateBookingInactiveEvent(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))
	inactive := &domain.Event{UUID: "6ba7b810-9dad-41d4-a716-446655440030", OwnerID: 1, Name: "Архивная", DurationMinutes: 30, IsActive: false}
	s.repos.AddEvent(inactive)
	status, data := s.do("POST", "/api/bookings", bookingBody(fake.BobUUID, inactive.UUID, "10:00"))
	assertError(t, status, data, 404, "not_found", "event not found")
}

func TestCreateBookingDSTGap(t *testing.T) {
	clock := &mutableClock{}
	s := newTestServer(t, clock)
	owner := fake.NewNYOwner(aliceUUID, "Alice Smith", "alice@example.com")
	s.repos.AddOwner(owner)
	ev := &domain.Event{UUID: "6ba7b810-9dad-41d4-a716-446655440020", OwnerID: owner.ID, Name: "Консультация", DurationMinutes: 30, IsActive: true}
	s.repos.AddEvent(ev)

	clock.Set(fake.NY(2026, time.March, 8, 0, 0))
	body := bookingBody(owner.UUID, ev.UUID, "02:30")
	body["date"] = "2026-03-08"
	status, data := s.do("POST", "/api/bookings", body)
	assertError(t, status, data, 400, "invalid_input", "ambiguous or nonexistent local time")
}

func TestCreateBookingDSTFold(t *testing.T) {
	clock := &mutableClock{}
	s := newTestServer(t, clock)
	owner := fake.NewNYOwner(aliceUUID, "Alice Smith", "alice@example.com")
	s.repos.AddOwner(owner)
	ev := &domain.Event{UUID: "6ba7b810-9dad-41d4-a716-446655440020", OwnerID: owner.ID, Name: "Консультация", DurationMinutes: 30, IsActive: true}
	s.repos.AddEvent(ev)

	clock.Set(fake.NY(2026, time.November, 1, 0, 0))
	body := bookingBody(owner.UUID, ev.UUID, "01:30")
	body["date"] = "2026-11-01"
	status, data := s.do("POST", "/api/bookings", body)
	assertError(t, status, data, 400, "invalid_input", "ambiguous or nonexistent local time")
}
