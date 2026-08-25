package usecase

import (
	"strconv"
	"strings"
	"time"
)

// loadLocation загружает IANA-таймзону; неизвестная зона → ErrInvalidTimezone.
func loadLocation(tz string) (*time.Location, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return nil, ErrInvalidTimezone
	}
	return loc, nil
}

// resolveLocalTime преобразует локальные дату и время в timezone владельца,
// отклоняя несуществующее (DST gap) и неоднозначное (DST fold) локальное время:
//
//   - gap: time.Date нормализует 02:30 → 01:30, wall-clock не совпадает;
//   - fold: Go выбирает первое вхождение (летнее), тогда t+1h даёт тот же
//     «wall clock» (зимнее время).
//
// Это поведение подтверждено эмпирически на Go (America/New_York, 2026).
func resolveLocalTime(loc *time.Location, y int, m time.Month, d, h, min int) (time.Time, error) {
	t := time.Date(y, m, d, h, min, 0, 0, loc)

	if !sameWallClock(t, y, m, d, h, min) {
		return time.Time{}, ErrAmbiguousTime
	}

	// Неоднозначность при переходе на зимнее время: соседнее UTC-мгновение
	// (±1 час) отображается в тот же локальный «wall clock».
	if sameWallClock(t.Add(time.Hour).In(loc), y, m, d, h, min) ||
		sameWallClock(t.Add(-time.Hour).In(loc), y, m, d, h, min) {
		return time.Time{}, ErrAmbiguousTime
	}

	return t, nil
}

func sameWallClock(t time.Time, y int, m time.Month, d, h, min int) bool {
	return t.Year() == y && t.Month() == m && t.Day() == d && t.Hour() == h && t.Minute() == min
}

// timeInLoc собирает локальное время из строки "HH:MM" на дату day в loc.
func timeInLoc(loc *time.Location, day time.Time, hhmm string) (time.Time, error) {
	parts := strings.Split(hhmm, ":")
	if len(parts) != 2 {
		return time.Time{}, ErrInvalidTime
	}
	h, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Time{}, ErrInvalidTime
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil {
		return time.Time{}, ErrInvalidTime
	}
	return time.Date(day.Year(), day.Month(), day.Day(), h, m, 0, 0, loc), nil
}

// weekdayKey возвращает короткий ключ дня недели ("mon".."sun").
func weekdayKey(t time.Time) string {
	return strings.ToLower(t.Weekday().String()[:3])
}

// dayMidnight возвращает полночь дня t в loc.
func dayMidnight(loc *time.Location, t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
}
