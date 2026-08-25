package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
)

// OwnerProvisioner — реализация repository.OwnerProvisioner на SQLite.
// Владелец и его default events создаются в одной транзакции (BEGIN IMMEDIATE).
type OwnerProvisioner struct {
	db *DB
}

// NewOwnerProvisioner создаёт provisioner.
func NewOwnerProvisioner(db *DB) *OwnerProvisioner {
	return &OwnerProvisioner{db: db}
}

// CreateOwnerWithDefaultEvents создаёт владельца и его default events атомарно.
// Поля OwnerID событий заполняются идентификатором созданного владельца.
func (p *OwnerProvisioner) CreateOwnerWithDefaultEvents(ctx context.Context, owner *domain.Owner, events []domain.Event) (int64, error) {
	var ownerID int64

	err := p.db.withinTx(ctx, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, `
			INSERT INTO owners
				(uuid, name, email, is_active, work_start, work_end, timezone, mon, tue, wed, thu, fri, sat, sun, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			owner.UUID,
			owner.Name,
			owner.Email,
			owner.IsActive,
			owner.WorkStart,
			owner.WorkEnd,
			owner.Timezone,
			dayColumn(owner.WorkingDays, "mon"),
			dayColumn(owner.WorkingDays, "tue"),
			dayColumn(owner.WorkingDays, "wed"),
			dayColumn(owner.WorkingDays, "thu"),
			dayColumn(owner.WorkingDays, "fri"),
			dayColumn(owner.WorkingDays, "sat"),
			dayColumn(owner.WorkingDays, "sun"),
			utcFormat(owner.CreatedAt),
		)
		if err != nil {
			return mapStoreError(err)
		}

		ownerID, err = res.LastInsertId()
		if err != nil {
			return fmt.Errorf("last insert id: %w", err)
		}

		for i := range events {
			events[i].OwnerID = ownerID
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO events (uuid, owner_id, name, description, duration_minutes, is_active, created_at)
				VALUES (?, ?, ?, ?, ?, ?, ?)`,
				events[i].UUID,
				events[i].OwnerID,
				events[i].Name,
				nullableString(events[i].Description),
				events[i].DurationMinutes,
				events[i].IsActive,
				utcFormat(events[i].CreatedAt),
			); err != nil {
				return mapStoreError(err)
			}
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	return ownerID, nil
}
