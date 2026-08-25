package sqlite

import (
	"context"
	"fmt"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
)

// EventRepository — реализация repository.EventRepository на SQLite.
type EventRepository struct {
	db *DB
}

// NewEventRepository создаёт репозиторий событий.
func NewEventRepository(db *DB) *EventRepository {
	return &EventRepository{db: db}
}

const eventColumns = `id, uuid, owner_id, name, description, duration_minutes, is_active, created_at`

// GetByOwnerID возвращает все события владельца, отсортированные по id.
func (r *EventRepository) GetByOwnerID(ctx context.Context, ownerID int64) ([]domain.Event, error) {
	rows, err := r.db.Conn().QueryContext(ctx, `SELECT `+eventColumns+` FROM events WHERE owner_id = ? ORDER BY id`, ownerID)
	if err != nil {
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	events := make([]domain.Event, 0)
	for rows.Next() {
		event, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate events: %w", err)
	}
	return events, nil
}

// GetByUUID возвращает событие по публичному UUID.
func (r *EventRepository) GetByUUID(ctx context.Context, uuid string) (*domain.Event, error) {
	row := r.db.Conn().QueryRowContext(ctx, `SELECT `+eventColumns+` FROM events WHERE uuid = ?`, uuid)
	event, err := scanEvent(row)
	if err != nil {
		return nil, mapStoreError(err)
	}
	return &event, nil
}

// GetByID возвращает событие по идентификатору.
func (r *EventRepository) GetByID(ctx context.Context, id int64) (*domain.Event, error) {
	row := r.db.Conn().QueryRowContext(ctx, `SELECT `+eventColumns+` FROM events WHERE id = ?`, id)
	event, err := scanEvent(row)
	if err != nil {
		return nil, mapStoreError(err)
	}
	return &event, nil
}

// Create вставляет событие и возвращает его идентификатор.
// Дубликат названия в рамках одного владельца отображается в domain.ErrConflict.
func (r *EventRepository) Create(ctx context.Context, event *domain.Event) (int64, error) {
	res, err := r.db.Conn().ExecContext(ctx, `
		INSERT INTO events (uuid, owner_id, name, description, duration_minutes, is_active, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		event.UUID,
		event.OwnerID,
		event.Name,
		nullableString(event.Description),
		event.DurationMinutes,
		event.IsActive,
		utcFormat(event.CreatedAt),
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

// Update обновляет название, описание, длительность и активность события.
func (r *EventRepository) Update(ctx context.Context, event *domain.Event) error {
	_, err := r.db.Conn().ExecContext(ctx, `
		UPDATE events SET name = ?, description = ?, duration_minutes = ?, is_active = ?
		WHERE id = ?`,
		event.Name,
		nullableString(event.Description),
		event.DurationMinutes,
		event.IsActive,
		event.ID,
	)
	if err != nil {
		return mapStoreError(err)
	}
	return nil
}

func scanEvent(s scanner) (domain.Event, error) {
	var (
		event      domain.Event
		descRaw    *string
		createdRaw string
	)
	if err := s.Scan(
		&event.ID,
		&event.UUID,
		&event.OwnerID,
		&event.Name,
		&descRaw,
		&event.DurationMinutes,
		&event.IsActive,
		&createdRaw,
	); err != nil {
		return domain.Event{}, err
	}

	if descRaw != nil {
		event.Description = *descRaw
	}

	createdAt, err := sqlTimeParse(createdRaw)
	if err != nil {
		return domain.Event{}, err
	}
	event.CreatedAt = createdAt

	return event, nil
}

// nullableString возвращает *string для NULL-колонки (description).
func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
