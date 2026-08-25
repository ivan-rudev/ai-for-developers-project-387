// Package usecase реализует бизнес-логику (use cases) приложения.
package usecase

import (
	"errors"
	"time"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/infrastructure/uuid"
)

// SlotWindowDays — размер окна слотов и бронирований: текущий день + следующие
// 13 дней (14 дней включительно), по docs/TESTING.md §0.3.
const SlotWindowDays = 14

// Clock — источник текущего времени. Позволяет детерминированно тестировать
// окно слотов и проверку «не в прошлом».
type Clock interface {
	Now() time.Time
}

// SystemClock возвращает реальное текущее время в UTC.
type SystemClock struct{}

// Now возвращает текущее время в UTC.
func (SystemClock) Now() time.Time { return time.Now().UTC() }

// UUIDGenerator — генератор UUID v4 для новых сущностей.
type UUIDGenerator interface {
	New() (string, error)
}

// UUIDV4 генерирует UUID v4, делегируя implementation infrastructure/uuid.
type UUIDV4 struct{}

// New возвращает новый UUID v4.
func (UUIDV4) New() (string, error) { return uuid.New() }

// Message возвращает пользовательское сообщение ошибки, если ошибка несёт его.
// Используется HTTP-слоем для формирования ErrorResponse.
func Message(err error) string {
	var me *messageError
	if errors.As(err, &me) {
		return me.message
	}
	return ""
}

type messageError struct {
	cause   error
	message string
}

func (e *messageError) Error() string { return e.message }

func (e *messageError) Unwrap() error { return e.cause }

// errMsg создаёт ошибку с пользовательским сообщением поверх доменной ошибки.
func errMsg(base error, message string) error {
	return &messageError{cause: base, message: message}
}
