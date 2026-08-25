// Package sqlite — реализация хранилища на SQLite. Содержит инициализацию БД,
// миграции (включая seed-данные), транзакции и репозитории.
//
// Поддерживается только режим одного экземпляра приложения с одним локальным
// файлом БД (single instance, один writer). Горизонтальное масштабирование
// не поддерживается.
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB — обёртка над соединением SQLite.
type DB struct {
	conn *sql.DB
}

// Open открывает (и при необходимости создаёт) базу по пути path,
// применяет миграции схемы и индексы. Драйвер настраивается через DSN:
//
//   - _txlock=immediate — все транзакции выполняются как BEGIN IMMEDIATE,
//     что защищает создание бронирований от гонок;
//   - _foreign_keys=on — включены внешние ключи;
//   - _busy_timeout — ожидание освобождения файла БД;
//   - _journal_mode=WAL — WAL-режим журнала.
func Open(path string) (*DB, error) {
	dsn := fmt.Sprintf("file:%s?_busy_timeout=5000&_journal_mode=WAL&_foreign_keys=on&_txlock=immediate", url.PathEscape(path))

	conn, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite %q: %w", path, err)
	}

	conn.SetMaxOpenConns(1)

	db := &DB{conn: conn}

	if err := db.Migrate(context.Background()); err != nil {
		return nil, errors.Join(err, conn.Close())
	}

	return db, nil
}

// Conn возвращает низкоуровневое соединение (используется репозиториями).
func (db *DB) Conn() *sql.DB {
	return db.conn
}

// Close закрывает соединение с базой данных.
func (db *DB) Close() error {
	return db.conn.Close()
}

// Migrate создаёт схему (таблицы и индексы), если она ещё не существует.
// Идемпотентно: повторные вызовы безопасны.
func (db *DB) Migrate(ctx context.Context) error {
	if _, err := db.conn.ExecContext(ctx, schemaSQL); err != nil {
		return fmt.Errorf("create schema: %w", err)
	}
	return nil
}

const schemaSQL = `
CREATE TABLE IF NOT EXISTS owners (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    work_start TEXT NOT NULL DEFAULT '09:00',
    work_end TEXT NOT NULL DEFAULT '18:00',
    timezone TEXT NOT NULL DEFAULT 'Europe/Moscow',
    mon BOOLEAN NOT NULL DEFAULT 1,
    tue BOOLEAN NOT NULL DEFAULT 1,
    wed BOOLEAN NOT NULL DEFAULT 1,
    thu BOOLEAN NOT NULL DEFAULT 1,
    fri BOOLEAN NOT NULL DEFAULT 1,
    sat BOOLEAN NOT NULL DEFAULT 0,
    sun BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS guests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    owner_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    duration_minutes INTEGER NOT NULL CHECK(duration_minutes > 0),
    is_active BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_id) REFERENCES owners(id) ON DELETE CASCADE,
    UNIQUE(owner_id, name)
);

CREATE TABLE IF NOT EXISTS bookings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_id INTEGER NOT NULL,
    guest_id INTEGER NOT NULL,
    event_id INTEGER NOT NULL,
    start_time DATETIME NOT NULL,
    end_time DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_id) REFERENCES owners(id) ON DELETE CASCADE,
    FOREIGN KEY (guest_id) REFERENCES guests(id) ON DELETE CASCADE,
    FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE RESTRICT,
    UNIQUE(owner_id, start_time, end_time)
);

CREATE INDEX IF NOT EXISTS idx_bookings_owner_id ON bookings(owner_id);
CREATE INDEX IF NOT EXISTS idx_bookings_event_id ON bookings(event_id);
CREATE INDEX IF NOT EXISTS idx_bookings_time_range ON bookings(owner_id, start_time, end_time);
CREATE INDEX IF NOT EXISTS idx_events_owner_id ON events(owner_id);
CREATE INDEX IF NOT EXISTS idx_events_uuid ON events(uuid);
CREATE INDEX IF NOT EXISTS idx_owners_uuid ON owners(uuid);
CREATE INDEX IF NOT EXISTS idx_owners_email ON owners(email);
CREATE INDEX IF NOT EXISTS idx_guests_email ON guests(email);
`

// withinTx выполняет fn внутри одной транзакции. За счёт DSN-параметра
// _txlock=immediate каждая транзакция выполняется как BEGIN IMMEDIATE:
// захватывается блокировка записи, что исключает параллельные записи.
func (db *DB) withinTx(ctx context.Context, fn func(tx *sql.Tx) error) (err error) {
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
			err = errors.Join(err, rbErr)
		}
	}()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// utcFormat сериализует time.Time в UTC для хранения в БД (ISO 8601, RFC 3339).
func utcFormat(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

// sqlTimeLayouts — возможные форматы DATETIME, встречающиеся в БД.
// Явные вставки выполняются в RFC 3339, DEFAULT CURRENT_TIMESTAMP — в "YYYY-MM-DD HH:MM:SS".
var sqlTimeLayouts = []string{
	time.RFC3339,
	"2006-01-02 15:04:05",
	"2006-01-02 15:04:05.999999999-07:00",
}

// sqlTimeParse разбирает значение DATETIME из БД в time.Time (UTC).
func sqlTimeParse(s string) (time.Time, error) {
	for _, layout := range sqlTimeLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("parse time %q: unsupported layout", s)
}
