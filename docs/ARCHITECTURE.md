# Architecture & Data Model

## MVP «Запись на звонок»

---

## 1. Архитектурный стиль

Проект реализуется по принципам **Clean Architecture** (Чистая Архитектура). Приоритет: независимость бизнес-логики от фреймворков, базы данных и UI.

### Правило зависимостей

Внутренние слои не зависят от внешних. Все стрелки зависимостей направлены к центру:

```
Infrastructure  →  Use Cases  →  Domain
       ↑                ↑
  Interfaces  →  Use Cases
```

### API-контракт

Интерфейсный контракт между фронтендом и бэкендом описывается на **TypeSpec** в директории `api/`. Сгенерированная OpenAPI 3.0 спецификация находится в `api/generated/openapi.yaml`.

- TypeSpec — единый источник правды для REST API.
- OpenAPI-спецификация генерируется командой `task gen:api` и коммитится в репозиторий.
- Слой **Interface Adapters** (`internal/interfaces/http/`) реализует эндпоинты, описанные в TypeSpec.
- Фронтенд (`web/js/api.js`) использует тот же контракт при интеграции с бэкендом.
- Дизайн-система фронтенда в стиле Cal.com описана в `docs/DESIGN.md` (палитра, типографика, компоненты, темы).

### Ограничения MVP

- MVP является учебным demo без аутентификации. Операции владельца и админские эндпоинты доступны без авторизации, а админка возвращает email гостей. Не разворачивать сервис в публичной или общей сети и не использовать реальные персональные данные.
- SQLite используется только в режиме одного экземпляра приложения с одним локальным volume базы данных. Горизонтальное масштабирование и общая файловая система не поддерживаются.
- Текущая конфигурация не добавляет runtime-ограничений на сетевую доступность: ответственность за изоляцию demo лежит на окружении запуска.

### Слои

| Слой | Ответственность | Пакет |
|------|-----------------|-------|
| **Domain** | Чистые структуры данных и бизнес-правила | `internal/domain` |
| **Use Cases** | Бизнес-логика: генерация слотов, создание бронирований, создание владельцев, управление событиями | `internal/usecase` |
| **Repository Ports** | Интерфейсы репозиториев, необходимые use case-слою | `internal/usecase/repository` |
| **Interface Adapters** | REST API хендлеры, роутер, middleware | `internal/interfaces/http` |
| **Infrastructure** | Конкретные реализации: SQLite, чтение YAML-конфига, rate limiter | `internal/infrastructure/sqlite`, `internal/infrastructure/config`, `internal/infrastructure/ratelimit` |

---

## 2. Структура папок

```
ai-for-developers-project-387/
├── api/
│   ├── main.tsp                  # TypeSpec: корневой файл спецификации
│   ├── models.tsp                # TypeSpec: модели
│   ├── errors.tsp                  # TypeSpec: ошибки
│   ├── routes/                   # TypeSpec: маршруты
│   └── generated/
│       └── openapi.yaml          # Сгенерированная OpenAPI 3.0 спецификация
├── cmd/
│   └── server/
│       └── main.go              # Точка входа: конфиг, DI, graceful shutdown
├── internal/
│   ├── domain/
│   │   ├── owner.go             # Owner
│   │   ├── guest.go             # Guest
│   │   ├── event.go             # Event
│   │   ├── booking.go           # Booking
│   │   ├── slot.go              # Slot
│   │   └── errors.go            # Доменные ошибки
│   ├── usecase/
│   │   ├── repository/          # Порты репозиториев (определяются use case-слоем)
│   │   │   ├── owner_repository.go
│   │   │   ├── guest_repository.go
│   │   │   ├── event_repository.go
│   │   │   └── booking_repository.go
│   │   ├── owner_usecase.go     # CRUD для владельцев
│   │   ├── event_usecase.go     # Управление событиями
│   │   ├── slot_usecase.go      # Генерация свободных слотов
│   │   ├── booking_usecase.go   # Создание бронирований, валидация
│   │   ├── admin_usecase.go     # Админка: default owner, предстоящие брони, создание событий
│   │   ├── owner_usecase_test.go
│   │   ├── event_usecase_test.go
│   │   ├── slot_usecase_test.go
│   │   ├── booking_usecase_test.go
│   │   └── admin_usecase_test.go
│   ├── interfaces/
│   │   └── http/
│   │       ├── router.go        # Маршрутизация, middleware
│   │       ├── owner_handler.go
│   │       ├── event_handler.go
│   │       ├── booking_handler.go
│   │       ├── slot_handler.go
│   │       ├── admin_handler.go   # GET /api/admin, /api/admin/bookings, /api/admin/events, POST /api/admin/events
│   │       ├── static_handler.go
│   │       └── middleware/
│   │           ├── logger.go
│   │           ├── recover.go
│   │           ├── cors.go
│   │           └── ratelimit.go
│   └── infrastructure/
│       ├── config/
│       │   └── config.go        # Чтение config.yaml
│       ├── ratelimit/
│       │   └── ratelimit.go     # In-memory rate limiter
│       └── sqlite/
│           ├── db.go            # Инициализация SQLite, миграции, транзакции
│           ├── owner_repository.go
│           ├── guest_repository.go
│           ├── event_repository.go
│           └── booking_repository.go
├── web/
│   ├── index.html               # SPA entry point
│   ├── fonts/                   # Cal Sans и Inter (OFL), локальная загрузка (docs/DESIGN.md)
│   ├── css/
│   │   ├── tokens.css           # Дизайн-токены: палитра, типографика, темы (docs/DESIGN.md)
│   │   └── styles.css           # Стили компонентов (docs/DESIGN.md)
│   └── js/
│       ├── app.js
│       ├── router.js
│       └── api.js
├── config.yaml                  # Глобальные настройки
├── Dockerfile                   # Multi-stage сборка
├── go.mod
├── go.sum
└── README.md
```

---

## 3. Data Model (SQLite)

### 3.1. Таблица `owners`

Хранит активных владельцев календарей. Настройки календаря встроены в ту же таблицу (MVP).

```sql
CREATE TABLE owners (
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
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | INTEGER PK | Внутренний автоинкремент, не отдаётся наружу |
| `uuid` | TEXT UNIQUE | Публичный UUID-идентификатор страницы владельца |
| `name` | TEXT | Имя владельца |
| `email` | TEXT UNIQUE | Email владельца |
| `is_active` | BOOLEAN | 1 = активный владелец, отображается в публичном списке |
| `work_start` | TEXT | Начало рабочего дня в часовом поясе `timezone`, `HH:MM` |
| `work_end` | TEXT | Окончание рабочего дня в часовом поясе `timezone`, `HH:MM` |
| `timezone` | TEXT | Часовой пояс для отображения слотов (IANA), например `Europe/Moscow` |
| `mon`..`sun` | BOOLEAN | Рабочий ли день недели |
| `created_at` | DATETIME | Время создания |

**Генерация `uuid`:** при создании владельца генерируется UUID v4 (например, `550e8400-e29b-41d4-a716-446655440000`). UUID не связан с PII и не выводится из name/email.

### 3.2. Таблица `guests`

Хранит гостей, которые забронировали слот. Гость не является владельцем до явного вызова `POST /api/owners`.

```sql
CREATE TABLE guests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | INTEGER PK | Внутренний автоинкремент |
| `name` | TEXT | Имя гостя |
| `email` | TEXT | Email гостя |
| `created_at` | DATETIME | Время создания |

Email в `guests` уникален — один и тот же email соответствует одному гостю. При создании бронирования сначала ищется существующий guest по email, иначе создаётся новый.

### 3.3. Таблица `events`

Хранит типы событий (event) владельца. Каждое событие определяет длительность встречи и заменяет собой уровень `slot_duration` у владельца.

```sql
CREATE TABLE events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT NOT NULL UNIQUE,
    owner_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    duration_minutes INTEGER NOT NULL CHECK(duration_minutes > 0),
    is_active BOOLEAN NOT NULL DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_id) REFERENCES owners(id) ON DELETE CASCADE,
    UNIQUE(owner_id, name)
);
```

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | INTEGER PK | Внутренний автоинкремент, не отдаётся наружу |
| `uuid` | TEXT UNIQUE | Публичный UUID-идентификатор события |
| `owner_id` | INTEGER FK | Владелец события |
| `name` | TEXT | Название события (уникально в рамках owner) |
| `description` | TEXT | Описание события |
| `duration_minutes` | INTEGER | Длительность события в минутах (например, 15 или 30) |
| `is_active` | BOOLEAN | 1 = активное событие, отображается в публичном списке |
| `created_at` | DATETIME | Время создания |

### 3.4. Таблица `bookings`

```sql
CREATE TABLE bookings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_id INTEGER NOT NULL,
    guest_id INTEGER NOT NULL,
    event_id INTEGER NOT NULL,
    start_time DATETIME NOT NULL,
    end_time DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_id) REFERENCES owners(id) ON DELETE CASCADE,
    FOREIGN KEY (guest_id) REFERENCES guests(id) ON DELETE CASCADE,
    FOREIGN KEY (event_id) REFERENCES events(id) ON DELETE RESTRICT,
    UNIQUE(owner_id, start_time, end_time)
);
```

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | INTEGER PK | Автоинкремент |
| `owner_id` | INTEGER FK | Владелец календаря |
| `guest_id` | INTEGER FK | Гость, создавший бронь |
| `event_id` | INTEGER FK | Выбранное событие (тип встречи) |
| `start_time` | DATETIME | Начало бронирования (UTC) |
| `end_time` | DATETIME | Окончание бронирования (UTC) |
| `created_at` | DATETIME | Время создания (UTC) |

**Защита от race condition:**
1. `UNIQUE(owner_id, start_time, end_time)` — не позволяет создать два идентичных бронирования.
2. Репозиторий `CreateBooking` выполняет `BEGIN IMMEDIATE` и внутри транзакции проверяет пересечение с существующими бронями перед вставкой.
3. Существующие бронирования влияют на генерацию слотов: пересекающиеся слоты помечаются как `unavailable` с `reason: "booked"`.

### 3.5. Индексы

```sql
CREATE INDEX idx_bookings_owner_id ON bookings(owner_id);
CREATE INDEX idx_bookings_event_id ON bookings(event_id);
CREATE INDEX idx_bookings_time_range ON bookings(owner_id, start_time, end_time);
CREATE INDEX idx_events_owner_id ON events(owner_id);
CREATE INDEX idx_events_uuid ON events(uuid);
CREATE INDEX idx_owners_uuid ON owners(uuid);
CREATE INDEX idx_owners_email ON owners(email);
CREATE INDEX idx_guests_email ON guests(email);
```

---

## 4. Конфигурация `config.yaml`

Файл находится в корне проекта. Используется как источник дефолтных настроек для новых владельцев и тестовых данных.

Порт HTTP-сервера можно переопределить переменной окружения `PORT` — она имеет
приоритет над `server.port` из файла (см. `cmd/server/main.go`, функция
`applyPortFromEnv`). Приоритет: `PORT` (env) > `config.yaml` > дефолт 8080.

```yaml
server:
  port: 8080
  host: "0.0.0.0"

health:
  path: "/healthz"

rate_limit:
  requests_per_minute: 30
  burst: 10

admin:
  owner_uuid: "550e8400-e29b-41d4-a716-446655440000"

default:
  work_start: "09:00"
  work_end: "18:00"
  timezone: "Europe/Moscow"
  working_days:
    - mon
    - tue
    - wed
    - thu
    - fri
  events:
    - name: "Короткая встреча"
      description: "Быстрый созвон на 15 минут"
      duration_minutes: 15
    - name: "Стандартная встреча"
      description: "Основная встреча на 30 минут"
      duration_minutes: 30

seed:
  owners:
    - uuid: "550e8400-e29b-41d4-a716-446655440000"
      name: "Bob Jones"
      email: "bob@example.com"
  events:
    - owner_uuid: "550e8400-e29b-41d4-a716-446655440000"
      name: "Короткая встреча"
      description: "Быстрый созвон на 15 минут"
      duration_minutes: 15
    - owner_uuid: "550e8400-e29b-41d4-a716-446655440000"
      name: "Стандартная встреча"
      description: "Основная встреча на 30 минут"
      duration_minutes: 30
```

### Описание секций

| Секция | Описание |
|--------|----------|
| `server` | Порт и хост для HTTP-сервера |
| `health` | Путь healthcheck endpoint |
| `rate_limit` | Базовый in-memory rate limiter |
| `admin` | UUID владельца, к которому привязана админская панель |
| `default` | Дефолтные настройки календаря для новых владельцев |
| `default.events` | Дефолтные события, автоматически создаваемые для каждого нового владельца (15 и 30 минут) |
| `seed` | Тестовые владельцы и события, создаваемые при миграции |

---

## 5. Доменные сущности (Go)

Доменный язык и термины (Owner, Guest, Event, Slot, Booking и др.) описаны в глоссарии [`CONTEXT.md`](CONTEXT.md); ниже — Go-структуры доменных сущностей.

```go
// internal/domain/owner.go
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

// internal/domain/guest.go
type Guest struct {
    ID        int64
    Name      string
    Email     string
    CreatedAt time.Time
}

// internal/domain/event.go
type Event struct {
    ID              int64
    UUID            string
    OwnerID         int64
    Name            string
    Description     string
    DurationMinutes int
    IsActive        bool
    CreatedAt       time.Time
}

// internal/domain/booking.go
type Booking struct {
    ID              int64
    OwnerID         int64
    GuestID         int64
    EventID         int64
    GuestName       string
    EventName       string
    GuestEmail      string
    DurationMinutes int
    StartTime       time.Time
    EndTime         time.Time
    CreatedAt       time.Time
}

// internal/domain/slot.go
type Slot struct {
    StartTime time.Time
    EndTime   time.Time
    Status    string // "available" | "unavailable"
    Reason    string // "booked" | "past", пусто для available
}
```

---

## 6. Контракты репозиториев (Use Case Ports)

Интерфейсы репозиториев определяются внутри слоя use case, чтобы не нарушать правило зависимостей.

```go
// internal/usecase/repository/owner_repository.go
type OwnerRepository interface {
    GetAll(ctx context.Context) ([]domain.Owner, error)
    GetByID(ctx context.Context, id int64) (*domain.Owner, error)
    GetByUUID(ctx context.Context, uuid string) (*domain.Owner, error)
    GetByEmail(ctx context.Context, email string) (*domain.Owner, error)
    Create(ctx context.Context, owner *domain.Owner) (int64, error)
    Update(ctx context.Context, owner *domain.Owner) error
}

// internal/usecase/repository/guest_repository.go
type GuestRepository interface {
    GetByID(ctx context.Context, id int64) (*domain.Guest, error)
    GetByEmail(ctx context.Context, email string) (*domain.Guest, error)
    Create(ctx context.Context, guest *domain.Guest) (int64, error)
}

// internal/usecase/repository/event_repository.go
type EventRepository interface {
    GetByOwnerID(ctx context.Context, ownerID int64) ([]domain.Event, error)
    GetByUUID(ctx context.Context, uuid string) (*domain.Event, error)
    GetByID(ctx context.Context, id int64) (*domain.Event, error)
    Create(ctx context.Context, event *domain.Event) (int64, error)
    Update(ctx context.Context, event *domain.Event) error
}

// internal/usecase/repository/booking_repository.go
type BookingRepository interface {
    GetByOwnerID(ctx context.Context, ownerID int64) ([]domain.Booking, error)
    GetByOwnerIDAndDate(ctx context.Context, ownerID int64, date time.Time) ([]domain.Booking, error)
    GetUpcomingByOwnerID(ctx context.Context, ownerID int64, from time.Time) ([]domain.Booking, error)
    CreateBooking(ctx context.Context, ownerID, guestID, eventID int64, start, end time.Time) (int64, error)
}

// internal/usecase/repository/owner_provisioner.go
// Выполняет owner и создание его default events в одной транзакции.
type OwnerProvisioner interface {
    CreateOwnerWithDefaultEvents(ctx context.Context, owner *domain.Owner, events []domain.Event) (int64, error)
}
```

`CreateBooking` отвечает за транзакционную проверку пересечений и вставку записи.
`OwnerProvisioner.CreateOwnerWithDefaultEvents` отвечает за атомарность создания
владельца и всех default events: при любой ошибке владелец не должен быть виден в
системе. `OwnerUsecase` использует этот порт вместо последовательных вызовов
`OwnerRepository.Create` и `EventRepository.Create`.

---

## 6.5. Админ Use Case

```go
// internal/usecase/admin_usecase.go
type AdminUsecase struct {
    ownerRepo   repository.OwnerRepository
    eventRepo   repository.EventRepository
    bookingRepo repository.BookingRepository
    adminUUID   string
    uuidGen     UUIDGenerator
    clock       Clock
}

// Возвращает владельца, UUID которого указан в config.yaml (admin.owner_uuid).
// В MVP это seed-владелец Bob.
func (uc *AdminUsecase) GetDefaultOwner(ctx context.Context) (*domain.Owner, error)

// Возвращает бронирования default owner с start_time >= now(), отсортированные по start_time.
func (uc *AdminUsecase) ListUpcomingBookings(ctx context.Context) ([]domain.Booking, error)

// Возвращает все события default owner (активные и неактивные).
func (uc *AdminUsecase) ListEvents(ctx context.Context) ([]domain.Event, error)

// Создаёт событие для default owner, генерируя UUID v4.
func (uc *AdminUsecase) CreateEvent(ctx context.Context, name, description string, durationMinutes int) (*domain.Event, error)
```

---

## 6.7. Команды разработки (Taskfile)

Канонический набор команд для локальной разработки и CI описан в `Taskfile.yaml` (go-task). Одна и та же команда работает и на машине разработчика, и в GitHub Actions.

| Команда | Назначение | Эквивалент |
|---------|------------|------------|
| `task setup` | Устанавливает инструменты в `bin/` (golangci-lint, gofumpt, gci) и npm-зависимости | `go install ...` + `npm ci` |
| `task gen:api` | Генерирует OpenAPI из TypeSpec | `npm run build:api` |
| `task watch:api` | Перегенерация OpenAPI в watch-режиме | `npm run watch:api` |
| `task check:api` | Проверяет, что `api/generated/openapi.yaml` синхронизирован с TypeSpec | `task gen:api` + `git diff --exit-code api/generated/openapi.yaml` |
| `task format` | Форматирует Go-код (gofumpt + gci) | `gofumpt -w` + `gci write` |
| `task build` | Собирает проект | `go build ./...` |
| `task run` | Запускает сервер локально | `go run ./cmd/server` |
| `task vet` | Статический анализ | `go vet ./...` |
| `task lint` | Линтинг | `golangci-lint run ./...` |
| `task test` | Юнит-тесты с race-детектором | `go test -v -race -count=1 ./...` |
| `task docker:build` | Собирает Docker-образ `calendar-mvp` | `docker build -t calendar-mvp .` |
| `task docker:run` | Запускает контейнер с volume `calendar-data` | `docker run -p 8080:8080 -v calendar-data:/app/data calendar-mvp` |
| `task ci` | Полная проверка перед коммитом | `task check:api` + `task vet` + `task lint` + `task test` |

---

## 7. Dockerfile

```dockerfile
# -----------------------------------------------------------------------------
# Stage 1: Go Backend Build
# -----------------------------------------------------------------------------
FROM golang:1.26-alpine AS backend

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o /bin/server ./cmd/server

# -----------------------------------------------------------------------------
# Stage 2: Final Runtime Image
# -----------------------------------------------------------------------------
FROM alpine:3.20

# tzdata: IANA-таймзоны для time.LoadLocation (Europe/Moscow и др.)
RUN apk add --no-cache ca-certificates sqlite-libs tzdata

WORKDIR /app

# Бинарник, конфиг и статика
COPY --from=backend /bin/server /app/server
COPY --from=backend /app/config.yaml /app/config.yaml
COPY --from=backend /app/web /app/web

# Рабочая директория для SQLite
RUN mkdir -p /app/data

# Порт HTTP-сервера (переопределяется через docker run -e PORT=9000)
ENV PORT=8080

EXPOSE ${PORT}

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT}/healthz || exit 1

CMD ["/app/server"]
```

### Запуск

```bash
task docker:build
task docker:run
```

Образ `calendar-mvp`; `task docker:run` запускает контейнер
`-p ${PORT}:${PORT} -e PORT=${PORT} -v calendar-data:/app/data` (порт по умолчанию 8080,
переопределяется через `PORT=9000 task docker:run`). `HEALTHCHECK` обращается к
`/healthz` на порту из `PORT`. Volume `calendar-data` сохраняет SQLite-базу между
перезапусками контейнера.

---

## 8. Graceful Shutdown

`cmd/server/main.go` запускает `http.Server` и обрабатывает сигналы `SIGINT`/`SIGTERM`
(упрощённо; адрес сервера на самом деле собирается из `cfg.Server.Host` и
`cfg.Server.Port`, где порт можно переопределить через env `PORT`):

```go
addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
srv := &http.Server{Addr: addr, Handler: router}

go func() {
    if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
        log.Fatalf("server error: %v", err)
    }
}()

quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
if err := srv.Shutdown(ctx); err != nil {
    log.Fatalf("server shutdown error: %v", err)
}
```

---

## 9. Доменные ошибки и HTTP-статусы

```go
// internal/domain/errors.go
var (
    ErrNotFound        = errors.New("not found")
    ErrConflict        = errors.New("conflict")
    ErrInvalidInput    = errors.New("invalid input")
    ErrSlotUnavailable = errors.New("slot unavailable")
    ErrOverlap         = errors.New("booking overlap")
    ErrRateLimit       = errors.New("rate limit exceeded")
)
```

| Доменная ошибка | HTTP-статус | Когда |
|-----------------|-------------|-------|
| `ErrNotFound` | 404 | Owner, guest, event или booking не найдены |
| `ErrConflict` | 409 | Email или название события уже заняты |
| `ErrInvalidInput` | 400 | Невалидный email, UUID, дата, время, длительность |
| `ErrSlotUnavailable` | 409 | Выбранный день выходной, слот не входит в рабочие часы или событие неактивно |
| `ErrOverlap` | 409 | Слот пересекается с существующим бронированием |
| `ErrRateLimit` | 429 | Слишком много запросов с одного IP |
| Остальные ошибки | 500 | Внутренние ошибки сервера |

---

## 10. Rate Limiting

In-memory rate limiter на базе `golang.org/x/time/rate`. Применяется к публичным мутациям: `POST /api/bookings`, `POST /api/owners`, `POST /api/owners/{uuid}/events` и `POST /api/admin/events`.

```go
// internal/infrastructure/ratelimit/ratelimit.go
limiter := rate.NewLimiter(rate.Every(time.Minute/30), 10)
```

Middleware возвращает `429 Too Many Requests` при превышении лимита.

---

## 10.5. Конвенции обработки ошибок, таймаутов и логирования

- **Обработка ошибок.** Все ошибки оборачиваются через `errors.Wrap()` с сохранением контекста вызова. Доменные ошибки из `internal/domain/errors.go` проверяются через `errors.Is()`.
- **Внешние вызовы.** Все внешние вызовы выполняются с таймаутом **3 секунды** и механизмом **retry**.
- **Логирование.** Используется `log/slog` на уровне **Info**.

---

## 11. Тестирование

### 11.1. Стратегия

- **TypeSpec/OpenAPI** — единый источник правды для HTTP-контрактов. При изменении API сначала обновляется `api/`, затем генерация `api/generated/openapi.yaml` через `task gen:api`, и только потом реализация в `internal/interfaces/http/`. Синхронизация сгенерированной спецификации с коммитом проверяется командой `task check:api` (используется в CI).
- Полный набор поведенческих тест-кейсов (Gherkin: API §1–7, конкурентный доступ §8, UI §9) — `docs/TESTING.md`; используется как источник сценариев для юнит- и HTTP-тестов.
- **Юнит-тесты** для слоя `usecase` — чистая бизнес-логика.
- **In-memory fake** реализации репозиториев в `internal/usecase/fake/`, включая `EventRepository`.
- Fake `BookingRepository.CreateBooking` должен имитировать транзакционную проверку пересечений и уникальность `(owner_id, start_time, end_time)`.
- **Clock provider** — `SlotUsecase` и `BookingUsecase` принимают интерфейс `Clock` для получения текущего времени. Это позволяет тестировать 14-дневное окно и прошедшие слоты детерминированно.
- Интеграционные тесты SQLite и HTTP-handlers для MVP не обязательны, но приветствуются.

### 11.2. Пример структуры тестов

```
internal/
└── usecase/
    ├── fake/
    │   ├── owner_repository.go
    │   ├── guest_repository.go
    │   ├── event_repository.go
    │   └── booking_repository.go
    ├── owner_usecase_test.go
    ├── event_usecase_test.go
    ├── slot_usecase_test.go
    └── booking_usecase_test.go
```

### 11.3. Пример теста генерации слотов

```go
func TestSlotUsecase_GenerateSlots(t *testing.T) {
    ownerRepo := &fake.OwnerRepository{
        Owners: map[string]*domain.Owner{
            "6ba7b810-9dad-41d4-a716-446655440000": {
                ID:          1,
                UUID:        "6ba7b810-9dad-41d4-a716-446655440000",
                Name:        "Bob",
                WorkStart:   "09:00",
                WorkEnd:     "18:00",
                Timezone:    "Europe/Moscow",
                WorkingDays: map[string]bool{"mon": true, "tue": true, "wed": true, "thu": true, "fri": true},
            },
        },
    }
    eventRepo := &fake.EventRepository{
        Events: map[string]*domain.Event{
            "event-uuid-30": {
                ID:              1,
                UUID:            "event-uuid-30",
                OwnerID:         1,
                Name:            "Стандартная встреча",
                DurationMinutes: 30,
                IsActive:        true,
            },
        },
    }
    bookingRepo := &fake.BookingRepository{Bookings: []domain.Booking{}}

    uc := usecase.NewSlotUsecase(ownerRepo, eventRepo, bookingRepo, fixedClock(date("2026-08-10")))
    slotsByDate, err := uc.GenerateSlots(context.Background(), "6ba7b810-9dad-41d4-a716-446655440000", "event-uuid-30")
    require.NoError(t, err)
    require.Equal(t, []string{"09:00", "09:30", "10:00", "10:30", "11:00"}, slotsByDate["2026-08-10"])
}
```

### 11.4. Запуск тестов

```bash
task test
```

Эквивалентная команда: `go test -v -race -count=1 ./...`.

---

## 12. Потоки данных

### 12.1. Список событий владельца

```
HTTP Request (owner_uuid)
    → EventHandler
    → OwnerUsecase / OwnerRepository (получить owner по uuid)
    → EventUsecase / EventRepository (получить события owner_id)
    → HTTP Response (events: uuid, name, description, duration_minutes, is_active)
```

### 12.2. Создание события

```
HTTP Request (owner_uuid, name, description, duration_minutes)
    → EventHandler
    → RateLimit middleware
    → OwnerUsecase / OwnerRepository (получить owner по uuid)
    → EventUsecase
        → EventRepository.Create (с UUID v4, owner_id, unique name per owner)
    → HTTP Response (uuid, name, description, duration_minutes, is_active)
```

### 12.3. Генерация слотов

```
HTTP Request (owner_uuid, event_uuid)
    → SlotHandler
    → OwnerUsecase / OwnerRepository (получить owner по uuid)
    → EventUsecase / EventRepository (получить событие по uuid)
    → SlotUsecase (owner.id, event.duration_minutes)
        → BookingRepository (получить брони за 14 дней)
    → SlotUsecase (вычислить все слоты на 14 дней в часовом поясе owner и пометить занятые/прошедшие как unavailable)
    → HTTP Response (slots grouped by date, each slot has time, status and optional reason)
```

### 12.4. Создание бронирования

```
HTTP Request (owner_uuid, event_uuid, guest_name, guest_email, date, start_time)
    → BookingHandler
    → RateLimit middleware
    → OwnerUsecase / OwnerRepository (получить owner по uuid)
    → EventUsecase / EventRepository (получить событие по uuid)
    → BookingUsecase
        → Валидация: email, длительность события, рабочий день, границы слота
        → GuestRepository (создать или найти guest по email)
        → BookingRepository.CreateBooking (транзакция: проверка пересечений + вставка)
    → Logger (mock email)
    → HTTP Response
```

### 12.5. Создание владельца (кнопка «Создать свой календарь»)

```
HTTP Request (name, email)
    → OwnerHandler
    → RateLimit middleware
    → OwnerUsecase
        → OwnerRepository.GetByEmail (проверка дубликата)
        → OwnerRepository.Create (с UUID v4 и дефолтными настройками)
        → EventRepository.Create (дефолтные события из config.yaml `default.events`)
    → HTTP Response (uuid, без email)
```

`OwnerUsecase` также зависит от `EventRepository`: после создания владельца автоматически создаются два события по умолчанию из `config.yaml` (`default.events`) — «Короткая встреча» (15 минут) и «Стандартная встреча» (30 минут), как у seed-владельца Bob.

### 12.6. Админ: предстоящие бронирования default owner

```
HTTP Request
    → AdminHandler
    → AdminUsecase
        → OwnerRepository.GetByUUID (admin.owner_uuid из конфига)
        → BookingRepository.GetUpcomingByOwnerID (owner_id, from = now())
    → HTTP Response (bookings: event_name, guest_name, guest_email, start_time, end_time, duration_minutes)
```

### 12.7. Админ: список событий default owner

```
HTTP Request
    → AdminHandler
    → AdminUsecase
        → OwnerRepository.GetByUUID (admin.owner_uuid из конфига)
        → EventRepository.GetByOwnerID (owner_id)
    → HTTP Response (events: uuid, name, description, duration_minutes, is_active)
```

### 12.8. Админ: создание события для default owner

```
HTTP Request (name, description, duration_minutes)
    → AdminHandler
    → RateLimit middleware
    → AdminUsecase
        → OwnerRepository.GetByUUID (admin.owner_uuid из конфига)
        → EventRepository.GetByOwnerID (проверка уникальности name)
        → EventRepository.Create (с UUID v4, owner_id, name, description, duration_minutes)
    → HTTP Response (uuid, name, description, duration_minutes, is_active)
```

---

## 13. Часовые пояса

1. Все времена в БД хранятся в UTC.
2. `work_start`/`work_end` задаются в часовом поясе `owner.Timezone`.
3. При генерации слотов определяется диапазон из 14 дней, начиная с текущей даты, в часовом поясе owner.
4. Длительность каждого слота равна `event.DurationMinutes`.
5. Каждая дата в диапазоне интерпретируется как полночь в часовом поясе owner.
6. Сгенерированные слоты возвращаются в локальном времени owner, сгруппированные по датам.
7. Frontend показывает локальное время owner (не время пользователя).
8. Для MVP допустимо, если frontend не конвертирует в часовой пояс пользователя, а отображает время в часовом поясе owner.
9. Форматы дат и времени на фронтенде — RU-локализация в часовом поясе owner (например, «пятница, 7 августа 2026 г.», «09:40–09:50»); подробнее в `docs/DESIGN.md` §5.
10. Часовой пояс владельца обязан быть валидным именем IANA. Некорректный идентификатор возвращает `400 Bad Request`.
11. Если комбинация локальных `date` и `start_time` попадает в DST-gap или является неоднозначной при переходе на зимнее время, создание бронирования отклоняется с `400 Bad Request`. Сервер не выбирает смещение молча.
