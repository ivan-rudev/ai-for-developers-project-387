package repository

import (
	"context"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
)

// OwnerProvisioner создаёт владельца вместе с его default events атомарно:
// либо владелец и все события появляются вместе, либо не появляется ничего.
// Реализация отвечает за транзакционность (один unit of work).
type OwnerProvisioner interface {
	CreateOwnerWithDefaultEvents(ctx context.Context, owner *domain.Owner, events []domain.Event) (int64, error)
}
