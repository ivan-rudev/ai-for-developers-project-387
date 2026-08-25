package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/infrastructure/config"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/infrastructure/uuid"
)

// Seed наполняет базу seed-данными из конфигурации и идемпотентно: повторный
// запуск не дублирует владельцев и события. Владельцы из cfg.Seed.Owners
// создаются с default events (cfg.Default.Events); события из cfg.Seed.Events
// добавляются владельцу, указанному в owner_uuid, если такого события ещё нет.
func (db *DB) Seed(ctx context.Context, cfg *config.Config) error {
	owners := NewOwnerRepository(db)
	events := NewEventRepository(db)
	provisioner := NewOwnerProvisioner(db)

	now := time.Now().UTC()

	if err := db.seedOwners(ctx, owners, provisioner, cfg, now); err != nil {
		return err
	}
	return db.seedEvents(ctx, owners, events, cfg, now)
}

// seedOwners создаёт владельцев из cfg.Seed.Owners с default events.
func (db *DB) seedOwners(ctx context.Context, owners *OwnerRepository, provisioner *OwnerProvisioner, cfg *config.Config, now time.Time) error {
	for _, so := range cfg.Seed.Owners {
		if err := db.seedOwner(ctx, owners, provisioner, so, cfg.Default, now); err != nil {
			return err
		}
	}
	return nil
}

// seedOwner создаёт одного владельца, если он ещё не существует.
func (db *DB) seedOwner(ctx context.Context, owners *OwnerRepository, provisioner *OwnerProvisioner, so config.SeedOwner, def config.Default, now time.Time) error {
	if _, err := owners.GetByEmail(ctx, so.Email); err == nil {
		return nil
	}

	ownerID := so.UUID
	if ownerID == "" {
		generated, err := uuid.New()
		if err != nil {
			return fmt.Errorf("seed owner %q: generate uuid: %w", so.Name, err)
		}
		ownerID = generated
	}

	defaultEvents, err := buildDefaultEvents(def.Events, now)
	if err != nil {
		return fmt.Errorf("seed owner %q: %w", so.Name, err)
	}

	owner := domain.Owner{
		UUID:        ownerID,
		Name:        so.Name,
		Email:       so.Email,
		IsActive:    true,
		WorkStart:   def.WorkStart,
		WorkEnd:     def.WorkEnd,
		Timezone:    def.Timezone,
		WorkingDays: toWorkingDays(def.WorkingDays),
		CreatedAt:   now,
	}

	if _, err := provisioner.CreateOwnerWithDefaultEvents(ctx, &owner, defaultEvents); err != nil {
		return fmt.Errorf("seed owner %q: %w", so.Name, err)
	}
	return nil
}

// seedEvents добавляет события из cfg.Seed.Events владельцам по owner_uuid.
func (db *DB) seedEvents(ctx context.Context, owners *OwnerRepository, events *EventRepository, cfg *config.Config, now time.Time) error {
	for _, se := range cfg.Seed.Events {
		owner, err := owners.GetByUUID(ctx, se.OwnerUUID)
		if err != nil {
			return fmt.Errorf("seed event %q: owner %q: %w", se.Name, se.OwnerUUID, err)
		}

		exists, err := eventExistsByName(ctx, events, owner.ID, se.Name)
		if err != nil {
			return fmt.Errorf("seed event %q: %w", se.Name, err)
		}
		if exists {
			continue
		}

		eventUUID, err := uuid.New()
		if err != nil {
			return fmt.Errorf("seed event %q: generate uuid: %w", se.Name, err)
		}

		if _, err := events.Create(ctx, &domain.Event{
			UUID:            eventUUID,
			OwnerID:         owner.ID,
			Name:            se.Name,
			Description:     se.Description,
			DurationMinutes: se.DurationMinutes,
			IsActive:        true,
			CreatedAt:       now,
		}); err != nil {
			return fmt.Errorf("seed event %q: %w", se.Name, err)
		}
	}

	return nil
}

// buildDefaultEvents превращает конфигурационные default events в доменные
// события с новыми UUID v4.
func buildDefaultEvents(cfg []config.EventConfig, createdAt time.Time) ([]domain.Event, error) {
	events := make([]domain.Event, 0, len(cfg))
	for _, e := range cfg {
		eventUUID, err := uuid.New()
		if err != nil {
			return nil, fmt.Errorf("default event %q: generate uuid: %w", e.Name, err)
		}
		events = append(events, domain.Event{
			UUID:            eventUUID,
			Name:            e.Name,
			Description:     e.Description,
			DurationMinutes: e.DurationMinutes,
			IsActive:        true,
			CreatedAt:       createdAt,
		})
	}
	return events, nil
}

// toWorkingDays преобразует список дней недели из конфигурации в map.
func toWorkingDays(days []string) map[string]bool {
	result := make(map[string]bool, len(days))
	for _, d := range days {
		result[d] = true
	}
	return result
}

func eventExistsByName(ctx context.Context, events *EventRepository, ownerID int64, name string) (bool, error) {
	list, err := events.GetByOwnerID(ctx, ownerID)
	if err != nil {
		return false, err
	}
	for _, e := range list {
		if e.Name == name {
			return true, nil
		}
	}
	return false, nil
}
