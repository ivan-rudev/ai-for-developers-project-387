package usecase

import (
	"context"
	"errors"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/usecase/repository"
)

// AdminUsecase — операции админской панели, привязанной к default owner
// (admin.owner_uuid из config.yaml, см. docs/TESTING.md §0.1).
type AdminUsecase struct {
	owners    repository.OwnerRepository
	events    repository.EventRepository
	bookings  repository.BookingRepository
	adminUUID string
	uuidGen   UUIDGenerator
	clock     Clock
}

// NewAdminUsecase создаёт AdminUsecase для владельца adminUUID.
func NewAdminUsecase(
	owners repository.OwnerRepository,
	events repository.EventRepository,
	bookings repository.BookingRepository,
	adminUUID string,
	uuidGen UUIDGenerator,
	clock Clock,
) *AdminUsecase {
	return &AdminUsecase{
		owners:    owners,
		events:    events,
		bookings:  bookings,
		adminUUID: adminUUID,
		uuidGen:   uuidGen,
		clock:     clock,
	}
}

// GetDefaultOwner возвращает default owner по admin.owner_uuid.
func (uc *AdminUsecase) GetDefaultOwner(ctx context.Context) (*domain.Owner, error) {
	return uc.owners.GetByUUID(ctx, uc.adminUUID)
}

// ListUpcomingBookings возвращает предстоящие бронирования default owner
// (start_time >= now), отсортированные по start_time.
func (uc *AdminUsecase) ListUpcomingBookings(ctx context.Context) ([]domain.Booking, error) {
	owner, err := uc.GetDefaultOwner(ctx)
	if err != nil {
		return nil, err
	}
	return uc.bookings.GetUpcomingByOwnerID(ctx, owner.ID, uc.clock.Now())
}

// ListEvents возвращает все события default owner, включая неактивные.
func (uc *AdminUsecase) ListEvents(ctx context.Context) ([]domain.Event, error) {
	owner, err := uc.GetDefaultOwner(ctx)
	if err != nil {
		return nil, err
	}
	return uc.events.GetByOwnerID(ctx, owner.ID)
}

// CreateEvent создаёт событие default owner с is_active=true и проверкой
// уникальности названия. Дубликат → ErrEventNameExists.
func (uc *AdminUsecase) CreateEvent(
	ctx context.Context,
	name, description string,
	durationMinutes int,
) (*domain.Event, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}
	if err := validateDuration(durationMinutes); err != nil {
		return nil, err
	}

	owner, err := uc.GetDefaultOwner(ctx)
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
