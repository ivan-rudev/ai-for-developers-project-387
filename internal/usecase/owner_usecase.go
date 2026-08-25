package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/usecase/repository"
)

// Defaults — дефолтные настройки и события для новых владельцев
// (соответствуют секции default конфигурации).
type Defaults struct {
	WorkStart   string
	WorkEnd     string
	Timezone    string
	WorkingDays []string
	Events      []DefaultEvent
}

// DefaultEvent — дефолтное событие нового владельца.
type DefaultEvent struct {
	Name            string
	Description     string
	DurationMinutes int
}

// OwnerUsecase — бизнес-логика владельцев календарей.
type OwnerUsecase struct {
	owners      repository.OwnerRepository
	bookings    repository.BookingRepository
	provisioner repository.OwnerProvisioner
	defaults    Defaults
	uuidGen     UUIDGenerator
	clock       Clock
}

// NewOwnerUsecase создаёт OwnerUsecase.
func NewOwnerUsecase(
	owners repository.OwnerRepository,
	bookings repository.BookingRepository,
	provisioner repository.OwnerProvisioner,
	defaults Defaults,
	uuidGen UUIDGenerator,
	clock Clock,
) *OwnerUsecase {
	return &OwnerUsecase{
		owners:      owners,
		bookings:    bookings,
		provisioner: provisioner,
		defaults:    defaults,
		uuidGen:     uuidGen,
		clock:       clock,
	}
}

// ListActiveOwners возвращает только активных владельцев.
func (uc *OwnerUsecase) ListActiveOwners(ctx context.Context) ([]domain.Owner, error) {
	all, err := uc.owners.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	active := make([]domain.Owner, 0, len(all))
	for _, owner := range all {
		if owner.IsActive {
			active = append(active, owner)
		}
	}
	return active, nil
}

// GetByUUID возвращает владельца по публичному UUID.
func (uc *OwnerUsecase) GetByUUID(ctx context.Context, uuid string) (*domain.Owner, error) {
	if !isValidUUIDV4(uuid) {
		return nil, ErrInvalidOwnerUUID
	}
	return uc.owners.GetByUUID(ctx, uuid)
}

// CreateOwner создаёт владельца и его default events атомарно через provisioner.
// Дубликат email → ErrEmailExists.
func (uc *OwnerUsecase) CreateOwner(ctx context.Context, name, email string) (*domain.Owner, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}
	if err := validateEmail(email); err != nil {
		return nil, err
	}

	if _, err := uc.owners.GetByEmail(ctx, email); err == nil {
		return nil, ErrEmailExists
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}

	ownerID, err := uc.uuidGen.New()
	if err != nil {
		return nil, fmt.Errorf("generate owner uuid: %w", err)
	}

	now := uc.clock.Now()
	owner := &domain.Owner{
		UUID:        ownerID,
		Name:        name,
		Email:       email,
		IsActive:    true,
		WorkStart:   uc.defaults.WorkStart,
		WorkEnd:     uc.defaults.WorkEnd,
		Timezone:    uc.defaults.Timezone,
		WorkingDays: toWorkingDays(uc.defaults.WorkingDays),
		CreatedAt:   now,
	}

	events := make([]domain.Event, 0, len(uc.defaults.Events))
	for _, de := range uc.defaults.Events {
		eventID, err := uc.uuidGen.New()
		if err != nil {
			return nil, fmt.Errorf("generate event uuid: %w", err)
		}
		events = append(events, domain.Event{
			UUID:            eventID,
			Name:            de.Name,
			Description:     de.Description,
			DurationMinutes: de.DurationMinutes,
			IsActive:        true,
			CreatedAt:       now,
		})
	}

	if _, err := uc.provisioner.CreateOwnerWithDefaultEvents(ctx, owner, events); err != nil {
		if errors.Is(err, domain.ErrConflict) {
			return nil, ErrEmailExists
		}
		return nil, err
	}

	return owner, nil
}

// ListBookings возвращает бронирования владельца, отсортированные по start_time.
func (uc *OwnerUsecase) ListBookings(ctx context.Context, ownerUUID string) ([]domain.Booking, error) {
	if !isValidUUIDV4(ownerUUID) {
		return nil, ErrInvalidOwnerUUID
	}
	owner, err := uc.owners.GetByUUID(ctx, ownerUUID)
	if err != nil {
		return nil, err
	}
	return uc.bookings.GetByOwnerID(ctx, owner.ID)
}

// toWorkingDays преобразует список дней недели в map-представление domain.Owner.
func toWorkingDays(days []string) map[string]bool {
	result := make(map[string]bool, len(days))
	for _, d := range days {
		result[d] = true
	}
	return result
}
