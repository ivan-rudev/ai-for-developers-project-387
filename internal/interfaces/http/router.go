package http

import (
	"log/slog"
	nethttp "net/http"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/infrastructure/ratelimit"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/interfaces/http/middleware"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/usecase"
)

// Dependencies — зависимости для сборки роутера.
type Dependencies struct {
	Owners      *usecase.OwnerUsecase
	Events      *usecase.EventUsecase
	Slots       *usecase.SlotUsecase
	Bookings    *usecase.BookingUsecase
	Admin       *usecase.AdminUsecase
	RateLimiter *ratelimit.Limiter
	Logger      *slog.Logger
	StaticDir   string
}

// NewRouter собирает HTTP-роутер со всеми маршрутами и middleware.
func NewRouter(deps Dependencies) nethttp.Handler {
	mux := nethttp.NewServeMux()

	health := NewHealthHandler()
	mux.HandleFunc("GET /healthz", health.Check)

	owners := NewOwnerHandler(deps.Owners)
	mux.HandleFunc("GET /api/owners", owners.List)
	mux.HandleFunc("POST /api/owners", owners.Create)
	mux.HandleFunc("GET /api/owners/{uuid}", owners.Get)
	mux.HandleFunc("GET /api/owners/{uuid}/bookings", owners.ListBookings)

	events := NewEventHandler(deps.Events)
	mux.HandleFunc("GET /api/owners/{uuid}/events", events.List)
	mux.HandleFunc("POST /api/owners/{uuid}/events", events.Create)

	slots := NewSlotHandler(deps.Owners, deps.Events, deps.Slots)
	mux.HandleFunc("GET /api/owners/{uuid}/slots", slots.List)

	bookings := NewBookingHandler(deps.Owners, deps.Bookings)
	mux.HandleFunc("POST /api/bookings", bookings.Create)

	admin := NewAdminHandler(deps.Admin)
	mux.HandleFunc("GET /api/admin", admin.GetOwner)
	mux.HandleFunc("GET /api/admin/bookings", admin.ListBookings)
	mux.HandleFunc("GET /api/admin/events", admin.ListEvents)
	mux.HandleFunc("POST /api/admin/events", admin.CreateEvent)

	mux.Handle("/", NewStaticHandler(deps.StaticDir))

	h := nethttp.Handler(mux)
	h = middleware.RateLimit(deps.RateLimiter)(h)
	h = middleware.CORS()(h)
	h = middleware.Logger(deps.Logger)(h)
	h = middleware.Recover(deps.Logger)(h)
	return h
}
