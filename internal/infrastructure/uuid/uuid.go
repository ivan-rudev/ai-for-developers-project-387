// Package uuid генерирует UUID v4 на базе криптографически безопасного
// генератора случайных чисел.
package uuid

import (
	"crypto/rand"
	"fmt"
)

// New возвращает новый UUID v4 в каноническом текстовом формате
// (например, 550e8400-e29b-41d4-a716-446655440000).
func New() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate uuid: %w", err)
	}

	// Версия 4 и вариант RFC 4122.
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
