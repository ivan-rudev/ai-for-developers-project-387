package middleware

import (
	"net"
	"net/http"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/infrastructure/ratelimit"
)

// RateLimit ограничивает частоту POST-запросов по IP клиента. GET-запросы не
// лимитируются (см. docs/TESTING.md §0.3).
func RateLimit(l *ratelimit.Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}
			if !l.Allow(clientIP(r)) {
				WriteError(w, http.StatusTooManyRequests, "rate_limit", "rate limit exceeded")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// clientIP извлекает IP из RemoteAddr, отбрасывая порт.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
