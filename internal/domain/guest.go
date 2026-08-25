package domain

import "time"

// Guest — гость, который забронировал слот. Гость не является владельцем
// до явного вызова создания владельца.
type Guest struct {
	ID        int64
	Name      string
	Email     string
	CreatedAt time.Time
}
