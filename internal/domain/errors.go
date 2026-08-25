package domain

import "errors"

var (
	// ErrNotFound - сущность (owner, guest, event, booking) не найдена.
	ErrNotFound = errors.New("not found")
	// ErrConflict - конфликт уникальности (email, название события).
	ErrConflict = errors.New("conflict")
	// ErrInvalidInput - невалидные входные данные.
	ErrInvalidInput = errors.New("invalid input")
	// ErrSlotUnavailable - слот недоступен (выходной, вне рабочих часов, событие неактивно).
	ErrSlotUnavailable = errors.New("slot unavailable")
	// ErrOverlap - пересечение с существующим бронированием.
	ErrOverlap = errors.New("booking overlap")
	// ErrRateLimit - превышен rate limit.
	ErrRateLimit = errors.New("rate limit exceeded")
)
