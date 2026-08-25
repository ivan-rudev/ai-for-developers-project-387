package domain

import "time"

// Owner — владелец календаря с настройками доступности.
// Настройки календаря встроены в сущность (MVP), поле SlotDuration отсутствует:
// длительность встречи определяется событиями (Event).
type Owner struct {
	ID          int64
	UUID        string
	Name        string
	Email       string
	IsActive    bool
	WorkStart   string // "HH:MM"
	WorkEnd     string // "HH:MM"
	Timezone    string // IANA, e.g. "Europe/Moscow"
	WorkingDays map[string]bool
	CreatedAt   time.Time
}

// IsWorkingDay возвращает true, если день недели является рабочим.
func (o *Owner) IsWorkingDay(day string) bool {
	if o.WorkingDays == nil {
		return false
	}
	return o.WorkingDays[day]
}
