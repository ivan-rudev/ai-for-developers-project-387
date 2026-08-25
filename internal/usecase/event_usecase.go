package usecase

import (
	"context"
	"errors"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/usecase/repository"
)

// EventUsecase — бизнес-логика событий владельца.
type EventUsecase struct {
	owners  repository.OwnerRepository
	events  repository.EventRepository
	uuidGen UUIDGenerator
	clock   Clock
}

// NewEventUsecase создаёт EventUsecase.
func NewEventUsecase(
	owners repository.OwnerRepository,
	events repository.EventRepository,
	uuidGen UUIDGenerator,
	clock Clock,
) *EventUsecase {
	return &EventUsecase{owners: owners, events: events, uuidGen: uuidGen, clock: clock}
}

// ListActiveByOwner возвращает активные события владельца.
func (uc *EventUsecase) ListActiveByOwner(ctx context.Context, ownerUUID string) ([]domain.Event, error) {
	if !isValidUUIDV4(ownerUUID) {
		return nil, ErrInvalidOwnerUUID
	}
	owner, err := uc.owners.GetByUUID(ctx, ownerUUID)
	if err != nil {
		return nil, err
	}
	list, err := uc.events.GetByOwnerID(ctx, owner.ID)
	if err != nil {
		return nil, err
	}
	active := make([]domain.Event, 0, len(list))
	for _, e := range list {
		if e.IsActive {
			active = append(active, e)
		}
	}
	return active, nil
}

// CreateForOwner создаёт событие владельца с is_active=true и проверкой
// уникальности названия в рамках владельца. Дубликат → ErrEventNameExists.
func (uc *EventUsecase) CreateForOwner(
	ctx context.Context,
	ownerUUID, name, description string,
	durationMinutes int,
) (*domain.Event, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}
	if err := validateDuration(durationMinutes); err != nil {
		return nil, err
	}
	if !isValidUUIDV4(ownerUUID) {
		return nil, ErrInvalidOwnerUUID
	}

	owner, err := uc.owners.GetByUUID(ctx, ownerUUID)
	if err != nil {
		return nil, err
	}

	list, err := uc.events.GetByOwnerID(ctx, owner.ID)
	if err != nil {
		return nil, err
	}
	for _, e := range list {
		if e.Name == name {
			return nil, ErrEventNameExists
		}
	}

	eventID, err := uc.uuidGen.New()
	if err != nil {
		return nil, err
	}
	event := &domain.Event{
		UUID:            eventID,
		OwnerID:         owner.ID,
		Name:            name,
		Description:     description,
		DurationMinutes: durationMinutes,
		IsActive:        true,
		CreatedAt:       uc.clock.Now(),
	}

	if _, err := uc.events.Create(ctx, event); err != nil {
		if errors.Is(err, domain.ErrConflict) {
			return nil, ErrEventNameExists
		}
		return nil, err
	}
	return event, nil
}

// GetByUUIDForOwner возвращает событие владельца по его UUID. Событие чужого
// владельца → domain.ErrNotFound.
func (uc *EventUsecase) GetByUUIDForOwner(ctx context.Context, ownerUUID, eventUUID string) (*domain.Event, error) {
	if !isValidUUIDV4(ownerUUID) {
		return nil, ErrInvalidOwnerUUID
	}
	if !isValidUUIDV4(eventUUID) {
		return nil, ErrInvalidEventUUID
	}
	owner, err := uc.owners.GetByUUID(ctx, ownerUUID)
	if err != nil {
		return nil, err
	}
	event, err := uc.events.GetByUUID(ctx, eventUUID)
	if err != nil {
		return nil, err
	}
	if event.OwnerID != owner.ID {
		return nil, domain.ErrNotFound
	}
	return event, nil
}
