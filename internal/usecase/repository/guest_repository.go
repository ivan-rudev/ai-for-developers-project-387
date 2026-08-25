package repository

import (
	"context"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
)

// GuestRepository — хранилище гостей.
type GuestRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.Guest, error)
	GetByEmail(ctx context.Context, email string) (*domain.Guest, error)
	Create(ctx context.Context, guest *domain.Guest) (int64, error)
}
