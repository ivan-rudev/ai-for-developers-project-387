package usecase

import "github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"

// Ошибки use case-слоя с человекочитаемыми сообщениями для ErrorResponse.
// Каждая оборачивает соответствующую доменную ошибку, которая определяет
// HTTP-статус (errors.Is(err, domain.Err*) работает через Unwrap).
var (
	// 400 Bad Request (domain.ErrInvalidInput).
	ErrInvalidEmail     = errMsg(domain.ErrInvalidInput, "invalid email")
	ErrNameRequired     = errMsg(domain.ErrInvalidInput, "name is required")
	ErrInvalidDuration  = errMsg(domain.ErrInvalidInput, "invalid duration")
	ErrInvalidOwnerUUID = errMsg(domain.ErrInvalidInput, "invalid owner uuid")
	ErrInvalidEventUUID = errMsg(domain.ErrInvalidInput, "invalid event uuid")
	ErrInvalidDate      = errMsg(domain.ErrInvalidInput, "invalid date")
	ErrInvalidTime      = errMsg(domain.ErrInvalidInput, "invalid time")
	ErrDateOutOfRange   = errMsg(domain.ErrInvalidInput, "date is outside available range")
	ErrAmbiguousTime    = errMsg(domain.ErrInvalidInput, "ambiguous or nonexistent local time")
	ErrInvalidTimezone  = errMsg(domain.ErrInvalidInput, "invalid timezone")
	ErrSlotOutsideHours = errMsg(domain.ErrInvalidInput, "slot is outside working hours")
	ErrSlotNotMultiple  = errMsg(domain.ErrInvalidInput, "slot start is not aligned with event duration")
	ErrSlotInPast       = errMsg(domain.ErrInvalidInput, "slot is in the past")
	ErrNotWorkingDay    = errMsg(domain.ErrInvalidInput, "selected day is not a working day")

	// 409 Conflict (domain.ErrConflict).
	ErrEmailExists     = errMsg(domain.ErrConflict, "email already exists")
	ErrEventNameExists = errMsg(domain.ErrConflict, "event name already exists")

	// 409 Conflict (domain.ErrOverlap).
	ErrOverlap = errMsg(domain.ErrOverlap, "booking overlap")
)
