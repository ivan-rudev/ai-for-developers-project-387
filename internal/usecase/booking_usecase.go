package usecase

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/usecase/repository"
)

// CreateBookingInput — входные данные для создания бронирования.
// Date и StartTime задаются в часовом поясе владельца.
type CreateBookingInput struct {
	OwnerUUID  string
	EventUUID  string
	GuestName  string
	GuestEmail string
	Date       string // "YYYY-MM-DD"
	StartTime  string // "HH:MM"
}

// BookingUsecase — создание и валидация бронирований.
type BookingUsecase struct {
	owners   repository.OwnerRepository
	guests   repository.GuestRepository
	events   repository.EventRepository
	bookings repository.BookingRepository
	logger   *slog.Logger
	clock    Clock
}

// NewBookingUsecase создаёт BookingUsecase. logger может быть nil — тогда
// используется slog.Default().
func NewBookingUsecase(
	owners repository.OwnerRepository,
	guests repository.GuestRepository,
	events repository.EventRepository,
	bookings repository.BookingRepository,
	logger *slog.Logger,
	clock Clock,
) *BookingUsecase {
	if logger == nil {
		logger = slog.Default()
	}
	return &BookingUsecase{
		owners:   owners,
		guests:   guests,
		events:   events,
		bookings: bookings,
		logger:   logger,
		clock:    clock,
	}
}

// CreateBooking выполняет полную валидацию и создаёт бронирование:
// формат входных данных → владелец/событие → окно 14 дней → рабочий день →
// рабочие часы → кратность длительности → «не в прошлом» → гость → транзакция.
func (uc *BookingUsecase) CreateBooking(ctx context.Context, in CreateBookingInput) (*domain.Booking, error) {
	if err := validateName(in.GuestName); err != nil {
		return nil, err
	}
	if err := validateEmail(in.GuestEmail); err != nil {
		return nil, err
	}
	if !isValidUUIDV4(in.OwnerUUID) {
		return nil, ErrInvalidOwnerUUID
	}
	if !isValidUUIDV4(in.EventUUID) {
		return nil, ErrInvalidEventUUID
	}
	date, err := time.Parse("2006-01-02", in.Date)
	if err != nil {
		return nil, ErrInvalidDate
	}
	if err := validateTimeHHMM(in.StartTime); err != nil {
		return nil, err
	}

	owner, err := uc.owners.GetByUUID(ctx, in.OwnerUUID)
	if err != nil {
		return nil, err
	}
	event, err := uc.events.GetByUUID(ctx, in.EventUUID)
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
	localStart, localEnd, err := validateSlot(owner, event, loc, date, in.StartTime, now)
	if err != nil {
		return nil, err
	}

	// Гость: поиск или создание.
	guest, err := uc.guests.GetByEmail(ctx, in.GuestEmail)
	if errors.Is(err, domain.ErrNotFound) {
		guestID, gerr := uc.guests.Create(ctx, &domain.Guest{
			Name:      in.GuestName,
			Email:     in.GuestEmail,
			CreatedAt: now,
		})
		if gerr != nil {
			return nil, gerr
		}
		guest = &domain.Guest{
			ID:        guestID,
			Name:      in.GuestName,
			Email:     in.GuestEmail,
			CreatedAt: now,
		}
	} else if err != nil {
		return nil, err
	}

	// Транзакционное создание с проверкой пересечений (ErrOverlap при конфликте).
	bookingID, err := uc.bookings.CreateBooking(ctx, owner.ID, guest.ID, event.ID, localStart.UTC(), localEnd.UTC())
	if err != nil {
		if errors.Is(err, domain.ErrOverlap) {
			return nil, ErrOverlap
		}
		return nil, err
	}

	booking := &domain.Booking{
		ID:              bookingID,
		OwnerID:         owner.ID,
		GuestID:         guest.ID,
		EventID:         event.ID,
		GuestName:       guest.Name,
		EventName:       event.Name,
		GuestEmail:      guest.Email,
		DurationMinutes: event.DurationMinutes,
		StartTime:       localStart.UTC(),
		EndTime:         localEnd.UTC(),
		CreatedAt:       now,
	}

	// Mock-отправка email владельцу.
	uc.logger.Info("mock email to owner",
		"owner_email", owner.Email,
		"guest_name", guest.Name,
		"guest_email", guest.Email,
		"event_name", event.Name,
		"start_time", booking.StartTime.Format(time.RFC3339),
	)

	return booking, nil
}

// validateSlot проверяет слот в часовом поясе владельца: окно 14 дней, рабочий
// день, рабочие часы, кратность длительности события и «не в прошлом».
// Возвращает локальные start/end слота.
func validateSlot(owner *domain.Owner, event *domain.Event, loc *time.Location, date time.Time, startTime string, now time.Time) (time.Time, time.Time, error) {
	h, min, err := parseHHMM(startTime)
	if err != nil {
		return time.Time{}, time.Time{}, ErrInvalidTime
	}

	localStart, err := resolveLocalTime(loc, date.Year(), date.Month(), date.Day(), h, min)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	today := dayMidnight(loc, now.In(loc))
	lastDay := today.AddDate(0, 0, SlotWindowDays-1)
	localDay := dayMidnight(loc, localStart)
	if localDay.Before(today) || localDay.After(lastDay) {
		return time.Time{}, time.Time{}, ErrDateOutOfRange
	}

	if !owner.IsWorkingDay(weekdayKey(localStart)) {
		return time.Time{}, time.Time{}, ErrNotWorkingDay
	}

	workStart, err := timeInLoc(loc, localDay, owner.WorkStart)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	workEnd, err := timeInLoc(loc, localDay, owner.WorkEnd)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	duration := time.Duration(event.DurationMinutes) * time.Minute
	localEnd := localStart.Add(duration)
	if localStart.Before(workStart) || localEnd.After(workEnd) {
		return time.Time{}, time.Time{}, ErrSlotOutsideHours
	}

	startMin := localStart.Hour()*60 + localStart.Minute()
	wsMin := workStart.Hour()*60 + workStart.Minute()
	if (startMin-wsMin)%event.DurationMinutes != 0 {
		return time.Time{}, time.Time{}, ErrSlotNotMultiple
	}

	if !localStart.After(now) {
		return time.Time{}, time.Time{}, ErrSlotInPast
	}

	return localStart, localEnd, nil
}

func parseHHMM(s string) (int, int, error) {
	parts := strings.Split(s, ":")
	h, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, err
	}
	return h, m, nil
}
