package usecase_test

import (
	"context"
	"regexp"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/usecase"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/usecase/fake"
)

var testCtx = context.Background()

var uuidV4Re = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func isUUIDv4(s string) bool { return uuidV4Re.MatchString(s) }

var testDefaults = usecase.Defaults{
	WorkStart:   "09:00",
	WorkEnd:     "18:00",
	Timezone:    "Europe/Moscow",
	WorkingDays: []string{"mon", "tue", "wed", "thu", "fri"},
	Events: []usecase.DefaultEvent{
		{Name: "Короткая встреча", DurationMinutes: 15},
		{Name: "Стандартная встреча", DurationMinutes: 30},
	},
}

func newOwnerUsecase(r *fake.Repositories) *usecase.OwnerUsecase {
	return usecase.NewOwnerUsecase(r.Owners, r.Bookings, r.Owners, testDefaults, fake.NewStubUUID(0), fake.NewFixedClock(fake.Today))
}

func newEventUsecase(r *fake.Repositories) *usecase.EventUsecase {
	return usecase.NewEventUsecase(r.Owners, r.Events, fake.NewStubUUID(0), fake.NewFixedClock(fake.Today))
}

func newSlotUsecase(r *fake.Repositories, clock usecase.Clock) *usecase.SlotUsecase {
	return usecase.NewSlotUsecase(r.Owners, r.Events, r.Bookings, clock)
}

func newBookingUsecase(r *fake.Repositories, clock usecase.Clock) *usecase.BookingUsecase {
	return usecase.NewBookingUsecase(r.Owners, r.Guests, r.Events, r.Bookings, nil, clock)
}

func newAdminUsecase(r *fake.Repositories, clock usecase.Clock) *usecase.AdminUsecase {
	return usecase.NewAdminUsecase(r.Owners, r.Events, r.Bookings, fake.BobUUID, fake.NewStubUUID(0), clock)
}

// eventByUUID возвращает событие из списка по UUID.
func eventByUUID(events []domain.Event, uuid string) *domain.Event {
	for i := range events {
		if events[i].UUID == uuid {
			return &events[i]
		}
	}
	return nil
}

// slotAt возвращает слот с заданным локальным временем "HH:MM".
func slotAt(slots []domain.Slot, hhmm string) domain.Slot {
	for _, s := range slots {
		if s.StartTime.Format("15:04") == hhmm {
			return s
		}
	}
	return domain.Slot{}
}
