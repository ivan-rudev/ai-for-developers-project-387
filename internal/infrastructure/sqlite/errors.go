package sqlite

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
)

// mapStoreError преобразует ошибку низкоуровневого хранилища в доменную ошибку.
// Уникальность email и названий событий обеспечивается UNIQUE-constraint'ами;
// sql.ErrNoRows преобразуется в ErrNotFound.
func mapStoreError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}

	if isUniqueViolation(err) {
		return domain.ErrConflict
	}

	return err
}

func isUniqueViolation(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "unique constraint failed")
}
