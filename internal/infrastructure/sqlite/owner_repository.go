package sqlite

import (
	"context"
	"fmt"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
)

// OwnerRepository — реализация repository.OwnerRepository на SQLite.
type OwnerRepository struct {
	db *DB
}

// NewOwnerRepository создаёт репозиторий владельцев.
func NewOwnerRepository(db *DB) *OwnerRepository {
	return &OwnerRepository{db: db}
}

const ownerColumns = `id, uuid, name, email, is_active, work_start, work_end, timezone, mon, tue, wed, thu, fri, sat, sun, created_at`

// GetAll возвращает всех владельцев, отсортированных по id.
func (r *OwnerRepository) GetAll(ctx context.Context) ([]domain.Owner, error) {
	rows, err := r.db.Conn().QueryContext(ctx, `SELECT `+ownerColumns+` FROM owners ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("list owners: %w", err)
	}
	defer rows.Close()

	owners := make([]domain.Owner, 0)
	for rows.Next() {
		owner, err := scanOwner(rows)
		if err != nil {
			return nil, err
		}
		owners = append(owners, owner)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate owners: %w", err)
	}
	return owners, nil
}

// GetByID возвращает владельца по внутреннему идентификатору.
func (r *OwnerRepository) GetByID(ctx context.Context, id int64) (*domain.Owner, error) {
	row := r.db.Conn().QueryRowContext(ctx, `SELECT `+ownerColumns+` FROM owners WHERE id = ?`, id)
	owner, err := scanOwner(row)
	if err != nil {
		return nil, mapStoreError(err)
	}
	return &owner, nil
}

// GetByUUID возвращает владельца по публичному UUID.
func (r *OwnerRepository) GetByUUID(ctx context.Context, uuid string) (*domain.Owner, error) {
	row := r.db.Conn().QueryRowContext(ctx, `SELECT `+ownerColumns+` FROM owners WHERE uuid = ?`, uuid)
	owner, err := scanOwner(row)
	if err != nil {
		return nil, mapStoreError(err)
	}
	return &owner, nil
}

// GetByEmail возвращает владельца по email.
func (r *OwnerRepository) GetByEmail(ctx context.Context, email string) (*domain.Owner, error) {
	row := r.db.Conn().QueryRowContext(ctx, `SELECT `+ownerColumns+` FROM owners WHERE email = ?`, email)
	owner, err := scanOwner(row)
	if err != nil {
		return nil, mapStoreError(err)
	}
	return &owner, nil
}

// Create вставляет владельца и возвращает его внутренний идентификатор.
// Дубликат email отображается в domain.ErrConflict.
func (r *OwnerRepository) Create(ctx context.Context, owner *domain.Owner) (int64, error) {
	res, err := r.db.Conn().ExecContext(ctx, `
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
		return 0, mapStoreError(err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}
	return id, nil
}

// Update обновляет настройки и статус владельца.
func (r *OwnerRepository) Update(ctx context.Context, owner *domain.Owner) error {
	_, err := r.db.Conn().ExecContext(ctx, `
		UPDATE owners SET
			name = ?, email = ?, is_active = ?, work_start = ?, work_end = ?, timezone = ?,
			mon = ?, tue = ?, wed = ?, thu = ?, fri = ?, sat = ?, sun = ?
		WHERE id = ?`,
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
		owner.ID,
	)
	if err != nil {
		return mapStoreError(err)
	}
	return nil
}

// scanner — интерфейс, реализуемый *sql.Row и *sql.Rows, чтобы переиспользовать
// логику чтения строки владельца.
type scanner interface {
	Scan(dest ...any) error
}

func scanOwner(s scanner) (domain.Owner, error) {
	var (
		owner      domain.Owner
		mon, tue   bool
		wed, thu   bool
		fri, sat   bool
		sun        bool
		createdRaw string
	)
	if err := s.Scan(
		&owner.ID,
		&owner.UUID,
		&owner.Name,
		&owner.Email,
		&owner.IsActive,
		&owner.WorkStart,
		&owner.WorkEnd,
		&owner.Timezone,
		&mon, &tue, &wed, &thu, &fri, &sat, &sun,
		&createdRaw,
	); err != nil {
		return domain.Owner{}, err
	}

	owner.WorkingDays = map[string]bool{
		"mon": mon,
		"tue": tue,
		"wed": wed,
		"thu": thu,
		"fri": fri,
		"sat": sat,
		"sun": sun,
	}

	createdAt, err := sqlTimeParse(createdRaw)
	if err != nil {
		return domain.Owner{}, err
	}
	owner.CreatedAt = createdAt

	return owner, nil
}

// dayColumn возвращает признак рабочего дня для колонки weekDay.
func dayColumn(days map[string]bool, weekDay string) bool {
	if days == nil {
		return false
	}
	return days[weekDay]
}
