package repository

import (
	"context"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
)

// EventRepository — хранилище событий владельцев.
type EventRepository interface {
	GetByOwnerID(ctx context.Context, ownerID int64) ([]domain.Event, error)
	GetByUUID(ctx context.Context, uuid string) (*domain.Event, error)
	GetByID(ctx context.Context, id int64) (*domain.Event, error)
	Create(ctx context.Context, event *domain.Event) (int64, error)
	Update(ctx context.Context, event *domain.Event) error
}
