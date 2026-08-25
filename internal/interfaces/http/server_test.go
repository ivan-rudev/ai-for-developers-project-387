package http

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	nethttp "net/http"
	"net/http/httptest"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/infrastructure/ratelimit"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/interfaces/http/middleware"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/usecase"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/usecase/fake"
)

// testServer — интегрированный HTTP-сервер на фейк-репозиториях и фиксированных часах.
type testServer struct {
	t     *testing.T
	srv   *httptest.Server
	repos *fake.Repositories
}

func newTestServer(t *testing.T, clock usecase.Clock) *testServer {
	t.Helper()
	return newTestServerWithLimiter(t, clock, ratelimit.New(1_000_000, 1_000_000))
}

func newTestServerWithLimiter(t *testing.T, clock usecase.Clock, limiter *ratelimit.Limiter) *testServer {
	t.Helper()
	repos, _, _ := fake.SeedBob()
	return buildTestServer(t, clock, limiter, repos)
}

func newEmptyTestServer(t *testing.T, clock usecase.Clock) *testServer {
	t.Helper()
	return buildTestServer(t, clock, ratelimit.New(1_000_000, 1_000_000), fake.NewRepositories())
}

func buildTestServer(t *testing.T, clock usecase.Clock, limiter *ratelimit.Limiter, repos *fake.Repositories) *testServer {
	t.Helper()

	defaults := usecase.Defaults{
		WorkStart:   "09:00",
		WorkEnd:     "18:00",
		Timezone:    "Europe/Moscow",
		WorkingDays: []string{"mon", "tue", "wed", "thu", "fri"},
		Events: []usecase.DefaultEvent{
			{Name: "Короткая встреча", Description: "Быстрый созвон на 15 минут", DurationMinutes: 15},
			{Name: "Стандартная встреча", Description: "Основная встреча на 30 минут", DurationMinutes: 30},
		},
	}

	uuidGen := fake.NewStubUUID(100)
	logger := slog.New(slog.DiscardHandler)

	owners := usecase.NewOwnerUsecase(repos.Owners, repos.Bookings, repos.Owners, defaults, uuidGen, clock)
	events := usecase.NewEventUsecase(repos.Owners, repos.Events, uuidGen, clock)
	slots := usecase.NewSlotUsecase(repos.Owners, repos.Events, repos.Bookings, clock)
	bookings := usecase.NewBookingUsecase(repos.Owners, repos.Guests, repos.Events, repos.Bookings, logger, clock)
	admin := usecase.NewAdminUsecase(repos.Owners, repos.Events, repos.Bookings, fake.BobUUID, uuidGen, clock)

	handler := NewRouter(Dependencies{
		Owners: owners, Events: events, Slots: slots, Bookings: bookings, Admin: admin,
		RateLimiter: limiter, Logger: logger, StaticDir: t.TempDir(),
	})

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return &testServer{t: t, srv: srv, repos: repos}
}

// do выполняет HTTP-запрос к тестовому серверу и возвращает статус и тело.
func (s *testServer) do(method, path string, body any) (int, []byte) {
	s.t.Helper()
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			s.t.Fatalf("marshal body: %v", err)
		}
		rdr = bytes.NewReader(b)
	}
	req, err := nethttp.NewRequest(method, s.srv.URL+path, rdr)
	if err != nil {
		s.t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := s.srv.Client().Do(req)
	if err != nil {
		s.t.Fatalf("do %s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		s.t.Fatalf("read body: %v", err)
	}
	return resp.StatusCode, data
}

// doJSON выполняет запрос и разбирает тело в v, возвращает статус.
func (s *testServer) doJSON(method, path string, body, v any) int {
	s.t.Helper()
	status, data := s.do(method, path, body)
	if v != nil {
		if err := json.Unmarshal(data, v); err != nil {
			s.t.Fatalf("decode %s %s: %v\nbody: %s", method, path, err, data)
		}
	}
	return status
}

func assertStatus(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("status = %d, want %d", got, want)
	}
}

// assertError проверяет статус, код и сообщение ошибки в теле ответа.
func assertError(t *testing.T, status int, data []byte, wantStatus int, code, message string) {
	t.Helper()
	assertStatus(t, status, wantStatus)
	var e middleware.ErrorResponse
	if err := json.Unmarshal(data, &e); err != nil {
		t.Fatalf("decode error body: %v\nbody: %s", err, data)
	}
	if e.Error != code {
		t.Fatalf("error code = %q, want %q", e.Error, code)
	}
	if e.Message != message {
		t.Fatalf("error message = %q, want %q", e.Message, message)
	}
}

// addAlice добавляет владельца Alice с событием «Консультация» (45 мин).
func (s *testServer) addAlice() (*domain.Owner, *domain.Event) {
	s.t.Helper()
	alice := fake.NewMoscowOwner("6ba7b810-9dad-41d4-a716-446655440010", "Alice Smith", "alice@example.com")
	s.repos.AddOwner(alice)
	ev := &domain.Event{
		UUID:            "6ba7b810-9dad-41d4-a716-446655440020",
		OwnerID:         alice.ID,
		Name:            "Консультация",
		Description:     "Индивидуальная консультация",
		DurationMinutes: 45,
		IsActive:        true,
	}
	s.repos.AddEvent(ev)
	return alice, ev
}

func TestDecodeJSONRejectsLargeRequest(t *testing.T) {
	t.Parallel()

	// Setup test server
	s := newTestServer(t, &mutableClock{})

	// Prepare a JSON body just above 1MB (1 + 1024 * 1024 bytes)
	largeSize := 1<<20 + 1
	largePayload := make(map[string]string)
	largePayload["data"] = string(make([]byte, largeSize))

	// Send POST request with large JSON
	status, data := s.do("POST", "/owners", largePayload)

	// Expect 400 Bad Request or similar
	if status == nethttp.StatusOK {
		t.Fatalf("expected error status for large JSON, got 200 OK")
	}

	// Decode error response
	var errResp struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(data, &errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	// Check error code presence
	if errResp.Error == "" {
		t.Fatalf("expected error code in response")
	}
}

// mutableClock — переключаемый источник времени для DST-тестов.
type mutableClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *mutableClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *mutableClock) Set(t time.Time) {
	c.mu.Lock()
	c.now = t
	c.mu.Unlock()
}

func jsonUnmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

var uuidV4Re = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func isValidUUID(s string) bool {
	return uuidV4Re.MatchString(s)
}
