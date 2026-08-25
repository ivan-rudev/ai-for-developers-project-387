package http

import (
	"net/http"
	"testing"
	"time"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/infrastructure/ratelimit"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/usecase/fake"
)

func TestRateLimitExceeded(t *testing.T) {
	s := newTestServerWithLimiter(t, fake.NewFixedClock(fake.Today), ratelimit.New(1, 2))

	for i, email := range []string{"alice@example.com", "bob2@example.com"} {
		status, _ := s.do("POST", "/api/owners", map[string]any{
			"name": "Alice", "email": email,
		})
		if status != http.StatusCreated {
			t.Fatalf("request %d status = %d, want %d", i, status, http.StatusCreated)
		}
	}

	status, data := s.do("POST", "/api/owners", map[string]any{
		"name": "Alice2", "email": "alice2@example.com",
	})
	assertError(t, status, data, http.StatusTooManyRequests, "rate_limit", "rate limit exceeded")
}

func TestRateLimitDoesNotApplyToGet(t *testing.T) {
	s := newTestServerWithLimiter(t, fake.NewFixedClock(fake.Today), ratelimit.New(1, 0))

	for i := range 20 {
		status, _ := s.do("GET", "/api/owners", nil)
		if status != http.StatusOK {
			t.Fatalf("GET %d status = %d, want %d", i, status, http.StatusOK)
		}
	}
}

func TestRateLimitRecovery(t *testing.T) {
	// 600 запросов/мин = 10 запросов в секунду; burst=1, поэтому после
	// первого запроса лимит исчерпан, но токен восполняется за ~100мс.
	s := newTestServerWithLimiter(t, fake.NewFixedClock(fake.Today), ratelimit.New(600, 1))

	status, _ := s.do("POST", "/api/owners", map[string]any{
		"name": "Alice", "email": "alice@example.com",
	})
	if status != http.StatusCreated {
		t.Fatalf("first request status = %d, want %d", status, http.StatusCreated)
	}

	status, data := s.do("POST", "/api/owners", map[string]any{
		"name": "Alice2", "email": "alice2@example.com",
	})
	assertError(t, status, data, http.StatusTooManyRequests, "rate_limit", "rate limit exceeded")

	time.Sleep(300 * time.Millisecond)

	status, _ = s.do("POST", "/api/owners", map[string]any{
		"name": "Alice3", "email": "alice3@example.com",
	})
	assertStatus(t, status, http.StatusCreated)
}
