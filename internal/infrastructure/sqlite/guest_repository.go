package sqlite

import (
	"context"
	"fmt"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
)

// GuestRepository — реализация repository.GuestRepository на SQLite.
type GuestRepository struct {
	db *DB
}

// NewGuestRepository создаёт репозиторий гостей.
func NewGuestRepository(db *DB) *GuestRepository {
	return &GuestRepository{db: db}
}

const guestColumns = `id, name, email, created_at`

// GetByID возвращает гостя по идентификатору.
func (r *GuestRepository) GetByID(ctx context.Context, id int64) (*domain.Guest, error) {
	row := r.db.Conn().QueryRowContext(ctx, `SELECT `+guestColumns+` FROM guests WHERE id = ?`, id)
	guest, err := scanGuest(row)
	if err != nil {
		return nil, mapStoreError(err)
	}
	return &guest, nil
}

// GetByEmail возвращает гостя по email.
func (r *GuestRepository) GetByEmail(ctx context.Context, email string) (*domain.Guest, error) {
	row := r.db.Conn().QueryRowContext(ctx, `SELECT `+guestColumns+` FROM guests WHERE email = ?`, email)
	guest, err := scanGuest(row)
	if err != nil {
		return nil, mapStoreError(err)
	}
	return &guest, nil
}

// Create вставляет гостя и возвращает его идентификатор.
func (r *GuestRepository) Create(ctx context.Context, guest *domain.Guest) (int64, error) {
	res, err := r.db.Conn().ExecContext(ctx,
		`INSERT INTO guests (name, email, created_at) VALUES (?, ?, ?)`,
		guest.Name, guest.Email, utcFormat(guest.CreatedAt),
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

func scanGuest(s scanner) (domain.Guest, error) {
	var (
		guest      domain.Guest
		createdRaw string
	)
	if err := s.Scan(&guest.ID, &guest.Name, &guest.Email, &createdRaw); err != nil {
		return domain.Guest{}, err
	}

	createdAt, err := sqlTimeParse(createdRaw)
	if err != nil {
		return domain.Guest{}, err
	}
	guest.CreatedAt = createdAt

	return guest, nil
}
