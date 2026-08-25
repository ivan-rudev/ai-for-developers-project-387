// Package repository определяет порты репозиториев, необходимые слою use case.
// Интерфейсы живут здесь, чтобы стрелка зависимости направлений шла к центру
// (Clean Architecture): use case зависит от абстракций, инфраструктура их реализует.
package repository

import (
	"context"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
)

// OwnerRepository — хранилище владельцев календарей.
type OwnerRepository interface {
	GetAll(ctx context.Context) ([]domain.Owner, error)
	GetByID(ctx context.Context, id int64) (*domain.Owner, error)
	GetByUUID(ctx context.Context, uuid string) (*domain.Owner, error)
	GetByEmail(ctx context.Context, email string) (*domain.Owner, error)
	Create(ctx context.Context, owner *domain.Owner) (int64, error)
	Update(ctx context.Context, owner *domain.Owner) error
}
