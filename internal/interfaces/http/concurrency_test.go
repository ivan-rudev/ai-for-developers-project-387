package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"testing"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/usecase/fake"
)

// concurrentDo выполняет HTTP-запрос к тестовому серверу без использования
// testing.T — безопасен для вызова из нескольких горутин.
func concurrentDo(s *testServer, method, path string, body any) (int, []byte, error) {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return 0, nil, err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, s.srv.URL+path, rdr)
	if err != nil {
		return 0, nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := s.srv.Client().Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}
	return resp.StatusCode, data, nil
}

// concurrentStatuses выполняет n запросов и возвращает набор HTTP-статусов.
func concurrentStatuses(t *testing.T, n int, fn func(i int) (int, []byte, error)) []int {
	t.Helper()
	statuses := make([]int, n)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			st, _, err := fn(i)
			if err != nil {
				st = 0
			}
			statuses[i] = st
		}(i)
	}
	close(start)
	wg.Wait()
	return statuses
}

func countStatuses(statuses []int, want int) int {
	n := 0
	for _, st := range statuses {
		if st == want {
			n++
		}
	}
	return n
}

// TC-8.1. Параллельное бронирование одного слота — ровно один 201, второй 409.
func TestConcurrentBookingSameSlot(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))

	statuses := concurrentStatuses(t, 2, func(i int) (int, []byte, error) {
		body := map[string]any{
			"owner_uuid":  fake.BobUUID,
			"event_uuid":  fake.StandardEventUUID,
			"guest_name":  "Jane Roe",
			"guest_email": "jane@example.com",
			"date":        "2026-08-12",
			"start_time":  "10:00",
		}
		if i == 1 {
			body["guest_name"] = "John Doe"
			body["guest_email"] = "john@example.com"
		}
		return concurrentDo(s, "POST", "/api/bookings", body)
	})

	if n := countStatuses(statuses, 201); n != 1 {
		t.Fatalf("201 responses = %d, want exactly 1 (statuses: %v)", n, statuses)
	}
	if n := countStatuses(statuses, 409); n != 1 {
		t.Fatalf("409 responses = %d, want exactly 1 (statuses: %v)", n, statuses)
	}
	if got := s.repos.Bookings.Count(); got != 1 {
		t.Fatalf("bookings count = %d, want 1", got)
	}
}

// TC-8.2. Параллельное бронирование пересекающихся слотов — ровно один 201.
// Валидные слоты одной длительности не пересекаются (границы кратны длительности
// от work_start), поэтому пересечение создаётся слотами разной длительности:
// «Стандартная встреча» 10:00–10:30 и «Короткая встреча» 10:15–10:30.
func TestConcurrentBookingOverlappingSlots(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))

	type req struct {
		eventUUID string
		startTime string
		email     string
	}
	reqs := []req{
		{eventUUID: fake.StandardEventUUID, startTime: "10:00", email: "jane@example.com"},
		{eventUUID: fake.ShortEventUUID, startTime: "10:15", email: "john@example.com"},
	}

	statuses := concurrentStatuses(t, len(reqs), func(i int) (int, []byte, error) {
		body := map[string]any{
			"owner_uuid":  fake.BobUUID,
			"event_uuid":  reqs[i].eventUUID,
			"guest_name":  "Jane Roe",
			"guest_email": reqs[i].email,
			"date":        "2026-08-12",
			"start_time":  reqs[i].startTime,
		}
		return concurrentDo(s, "POST", "/api/bookings", body)
	})

	if n := countStatuses(statuses, 201); n != 1 {
		t.Fatalf("201 responses = %d, want exactly 1 (statuses: %v)", n, statuses)
	}
	if n := countStatuses(statuses, 409); n != 1 {
		t.Fatalf("409 responses = %d, want exactly 1 (statuses: %v)", n, statuses)
	}
	if got := s.repos.Bookings.Count(); got != 1 {
		t.Fatalf("bookings count = %d, want 1", got)
	}
}

// TC-8.3. Параллельное создание владельцев с одним email — ровно один 201.
func TestConcurrentCreateOwnerSameEmail(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))

	statuses := concurrentStatuses(t, 2, func(i int) (int, []byte, error) {
		return concurrentDo(s, "POST", "/api/owners", map[string]any{
			"name":  "Dup Owner",
			"email": "dup@example.com",
		})
	})

	if n := countStatuses(statuses, 201); n != 1 {
		t.Fatalf("201 responses = %d, want exactly 1 (statuses: %v)", n, statuses)
	}
	if n := countStatuses(statuses, 409); n != 1 {
		t.Fatalf("409 responses = %d, want exactly 1 (statuses: %v)", n, statuses)
	}

	ctx := context.Background()
	owner, err := s.repos.Owners.GetByEmail(ctx, "dup@example.com")
	if err != nil {
		t.Fatalf("owner with dup@example.com not found: %v", err)
	}
	events, err := s.repos.Events.GetByOwnerID(ctx, owner.ID)
	if err != nil {
		t.Fatalf("get events: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("default events count = %d, want 2", len(events))
	}
	for _, e := range events {
		if e.DurationMinutes != 15 && e.DurationMinutes != 30 {
			t.Fatalf("unexpected default event duration: %d", e.DurationMinutes)
		}
	}
}

// TC-8.4. Параллельное создание событий с одним названием — ровно один 201.
func TestConcurrentCreateEventSameName(t *testing.T) {
	s := newTestServer(t, fake.NewFixedClock(fake.Today))

	statuses := concurrentStatuses(t, 2, func(i int) (int, []byte, error) {
		return concurrentDo(s, "POST", "/api/owners/"+fake.BobUUID+"/events", map[string]any{
			"name":             "Консультация",
			"description":      "Индивидуальная консультация",
			"duration_minutes": 45,
		})
	})

	if n := countStatuses(statuses, 201); n != 1 {
		t.Fatalf("201 responses = %d, want exactly 1 (statuses: %v)", n, statuses)
	}
	if n := countStatuses(statuses, 409); n != 1 {
		t.Fatalf("409 responses = %d, want exactly 1 (statuses: %v)", n, statuses)
	}

	ctx := context.Background()
	bob, err := s.repos.Owners.GetByUUID(ctx, fake.BobUUID)
	if err != nil {
		t.Fatalf("get bob: %v", err)
	}
	events, err := s.repos.Events.GetByOwnerID(ctx, bob.ID)
	if err != nil {
		t.Fatalf("get events: %v", err)
	}
	n := 0
	for _, e := range events {
		if e.Name == "Консультация" {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("events named «Консультация» count = %d, want 1", n)
	}
}
