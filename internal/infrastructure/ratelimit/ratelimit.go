// Package ratelimit реализует in-memory rate limiter для HTTP API.
// Лимит применяется по ключу (обычно IP клиента) с поддержкой burst.
package ratelimit

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	// cleanupInterval — периодичность очистки неактивных записей.
	cleanupInterval = 10 * time.Minute
	// entryTTL — время бездействия, после которого запись удаляется.
	entryTTL = 30 * time.Minute
)

// Limiter — in-memory rate limiter, ограничивающий частоту запросов по ключу
// (например, IP клиента). Значение "запросов в минуту" задаётся при создании.
type Limiter struct {
	mu      sync.Mutex
	entries map[string]*entry
	limit   rate.Limit
	burst   int
}

type entry struct {
	limiter *rate.Limiter
	last    time.Time
}

// New создаёт Limiter с requestsPerMinute запросов в минуту и burst-размером
// всплеска. Значения меньше 1 заменяются на 1.
func New(requestsPerMinute, burst int) *Limiter {
	if requestsPerMinute < 1 {
		requestsPerMinute = 1
	}
	if burst < 1 {
		burst = 1
	}
	l := &Limiter{
		entries: make(map[string]*entry),
		limit:   rate.Every(time.Minute / time.Duration(requestsPerMinute)),
		burst:   burst,
	}
	go l.cleanup()
	return l
}

// Allow возвращает true, если запрос для ключа key разрешён.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	e, ok := l.entries[key]
	if !ok {
		e = &entry{limiter: rate.NewLimiter(l.limit, l.burst), last: now}
		l.entries[key] = e
	}
	e.last = now
	return e.limiter.Allow()
}

// cleanup периодически удаляет записи, не использованные дольше entryTTL.
func (l *Limiter) cleanup() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()
	for range ticker.C {
		l.mu.Lock()
		for key, e := range l.entries {
			if time.Since(e.last) > entryTTL {
				delete(l.entries, key)
			}
		}
		l.mu.Unlock()
	}
}
