// Package main — точка входа HTTP-сервера приложения.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/infrastructure/config"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/infrastructure/ratelimit"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/infrastructure/sqlite"
	apphttp "github.com/ivan-rudev/ai-for-developers-project-387/internal/interfaces/http"
	"github.com/ivan-rudev/ai-for-developers-project-387/internal/usecase"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("server exited with error", "error", err)
		os.Exit(1)
	}
}

// run настраивает зависимости, запускает HTTP-сервер и дожидается
// сигнала остановки для graceful shutdown.
func run(logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load(getenv("CONFIG_PATH", "config.yaml"))
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if err := applyPortFromEnv(cfg, os.Getenv); err != nil {
		return err
	}

	dbPath := getenv("DB_PATH", "data/calendar.db")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o750); err != nil {
		return fmt.Errorf("create data directory: %w", err)
	}

	db, err := sqlite.Open(dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Seed(ctx, cfg); err != nil {
		return fmt.Errorf("seed database: %w", err)
	}

	ownersRepo := sqlite.NewOwnerRepository(db)
	guestsRepo := sqlite.NewGuestRepository(db)
	eventsRepo := sqlite.NewEventRepository(db)
	bookingsRepo := sqlite.NewBookingRepository(db)
	provisioner := sqlite.NewOwnerProvisioner(db)

	clock := usecase.SystemClock{}
	uuidGen := usecase.UUIDV4{}

	owners := usecase.NewOwnerUsecase(ownersRepo, bookingsRepo, provisioner, defaultsFrom(cfg), uuidGen, clock)
	events := usecase.NewEventUsecase(ownersRepo, eventsRepo, uuidGen, clock)
	slots := usecase.NewSlotUsecase(ownersRepo, eventsRepo, bookingsRepo, clock)
	bookings := usecase.NewBookingUsecase(ownersRepo, guestsRepo, eventsRepo, bookingsRepo, logger, clock)
	admin := usecase.NewAdminUsecase(ownersRepo, eventsRepo, bookingsRepo, cfg.Admin.OwnerUUID, uuidGen, clock)

	limiter := ratelimit.New(cfg.RateLimit.RequestsPerMinute, cfg.RateLimit.Burst)

	handler := apphttp.NewRouter(apphttp.Dependencies{
		Owners:      owners,
		Events:      events,
		Slots:       slots,
		Bookings:    bookings,
		Admin:       admin,
		RateLimiter: limiter,
		Logger:      logger,
		StaticDir:   getenv("STATIC_DIR", "web"),
	})

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("server listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		logger.Info("shutting down", "reason", ctx.Err().Error())
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}

// defaultsFrom превращает конфигурационные default-настройки в use case Defaults.
func defaultsFrom(cfg *config.Config) usecase.Defaults {
	d := usecase.Defaults{
		WorkStart:   cfg.Default.WorkStart,
		WorkEnd:     cfg.Default.WorkEnd,
		Timezone:    cfg.Default.Timezone,
		WorkingDays: cfg.Default.WorkingDays,
	}
	for _, e := range cfg.Default.Events {
		d.Events = append(d.Events, usecase.DefaultEvent{
			Name:            e.Name,
			Description:     e.Description,
			DurationMinutes: e.DurationMinutes,
		})
	}
	return d
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// applyPortFromEnv переопределяет cfg.Server.Port значением переменной окружения
// PORT, если она задана. Приоритет: PORT > config.yaml > дефолт 8080.
func applyPortFromEnv(cfg *config.Config, getenv func(string) string) error {
	if v := getenv("PORT"); v != "" {
		port, err := resolvePort(v)
		if err != nil {
			return err
		}
		cfg.Server.Port = port
	}
	return nil
}

// resolvePort парсит значение PORT как номер порта в диапазоне 1–65535.
func resolvePort(v string) (int, error) {
	port, err := strconv.Atoi(v)
	if err != nil || port < 1 || port > 65535 {
		return 0, fmt.Errorf("invalid PORT %q: must be an integer between 1 and 65535", v)
	}
	return port, nil
}
