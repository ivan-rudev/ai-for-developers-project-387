package usecase

import (
	"context"
	"time"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/usecase/repository"
)

// SlotUsecase — генерация слотов владельца на окно бронирования.
type SlotUsecase struct {
	owners   repository.OwnerRepository
	events   repository.EventRepository
	bookings repository.BookingRepository
	clock    Clock
}

// NewSlotUsecase создаёт SlotUsecase.
func NewSlotUsecase(
	owners repository.OwnerRepository,
	events repository.EventRepository,
	bookings repository.BookingRepository,
	clock Clock,
) *SlotUsecase {
	return &SlotUsecase{owners: owners, events: events, bookings: bookings, clock: clock}
}

// GenerateSlots возвращает слоты события на окно в SlotWindowDays дней (сегодня
// включительно), сгруппированные по дате "YYYY-MM-DD" в часовом поясе владельца.
// Выходные дни присутствуют с пустым списком слотов. Прошедшие слоты текущего
// дня помечаются unavailable/reason=past, пересекающиеся с бронями —
// unavailable/reason=booked.
//
// Событие должно принадлежать владельцу и быть активным, иначе → domain.ErrNotFound.
func (uc *SlotUsecase) GenerateSlots(ctx context.Context, ownerUUID, eventUUID string) (map[string][]domain.Slot, error) {
	if !isValidUUIDV4(ownerUUID) {
		return nil, ErrInvalidOwnerUUID
	}
	if !isValidUUIDV4(eventUUID) {
		return nil, ErrInvalidEventUUID
	}

	owner, err := uc.owners.GetByUUID(ctx, ownerUUID)
	if err != nil {
		return nil, err
	}
	event, err := uc.events.GetByUUID(ctx, eventUUID)
	if err != nil {
		return nil, err
	}
	if event.OwnerID != owner.ID || !event.IsActive {
		return nil, domain.ErrNotFound
	}

	loc, err := loadLocation(owner.Timezone)
	if err != nil {
		return nil, err
	}

	now := uc.clock.Now()
	today := dayMidnight(loc, now.In(loc))

	bookings, err := uc.bookings.GetByOwnerID(ctx, owner.ID)
	if err != nil {
		return nil, err
	}
	windowBookings := filterBookingsInWindow(bookings, today, SlotWindowDays)

	duration := time.Duration(event.DurationMinutes) * time.Minute
	result := make(map[string][]domain.Slot, SlotWindowDays)

	for i := range SlotWindowDays {
		day := today.AddDate(0, 0, i)
		key := day.Format("2006-01-02")

		if !owner.IsWorkingDay(weekdayKey(day)) {
			result[key] = []domain.Slot{}
			continue
		}

		workStart, err := timeInLoc(loc, day, owner.WorkStart)
		if err != nil {
			return nil, err
		}
		workEnd, err := timeInLoc(loc, day, owner.WorkEnd)
		if err != nil {
			return nil, err
		}

		slots := make([]domain.Slot, 0)
		for start := workStart; !start.Add(duration).After(workEnd); start = start.Add(duration) {
			slot := domain.Slot{
				StartTime: start,
				EndTime:   start.Add(duration),
				Status:    domain.SlotStatusAvailable,
			}

			switch {
			case !slot.StartTime.After(now):
				slot.Status = domain.SlotStatusUnavailable
				slot.Reason = domain.UnavailableReasonPast
			case overlapsAny(slot, windowBookings):
				slot.Status = domain.SlotStatusUnavailable
				slot.Reason = domain.UnavailableReasonBooked
			}

			slots = append(slots, slot)
		}
		result[key] = slots
	}

	return result, nil
}

// BookingWindow возвращает даты начала и конца окна бронирования ("YYYY-MM-DD")
// в часовом поясе владельца: [сегодня, сегодня + SlotWindowDays - 1].
func (uc *SlotUsecase) BookingWindow(ctx context.Context, ownerUUID string) (startDate, endDate string, err error) {
	owner, err := uc.owners.GetByUUID(ctx, ownerUUID)
	if err != nil {
		return "", "", err
	}
	loc, err := loadLocation(owner.Timezone)
	if err != nil {
		return "", "", err
	}
	today := dayMidnight(loc, uc.clock.Now().In(loc))
	return today.Format("2006-01-02"), today.AddDate(0, 0, SlotWindowDays-1).Format("2006-01-02"), nil
}

// filterBookingsInWindow оставляет брони, пересекающие окно [windowStart, +days).
func filterBookingsInWindow(bookings []domain.Booking, windowStart time.Time, days int) []domain.Booking {
	start := windowStart.UTC()
	end := windowStart.AddDate(0, 0, days).UTC()
	filtered := make([]domain.Booking, 0, len(bookings))
	for _, b := range bookings {
		if b.StartTime.Before(end) && b.EndTime.After(start) {
			filtered = append(filtered, b)
		}
	}
	return filtered
}

// overlapsAny проверяет пересечение слота с любой бронью (полуинтервалы).
func overlapsAny(slot domain.Slot, bookings []domain.Booking) bool {
	for _, b := range bookings {
		if slot.StartTime.Before(b.EndTime) && slot.EndTime.After(b.StartTime) {
			return true
		}
	}
	return false
}
