# Step-by-Step Todo List

## MVP «Запись на звонок» — Spec-Driven Development

---

## Фаза 1: Подготовка и структура проекта

- [x] Прочитать и проанализировать `CONSTITUTION.md`
- [x] Уточнить требования с пользователем (фронтенд, данные, диапазон, формат)
- [x] Создать `docs/PRD.md`
- [x] Создать `docs/ARCHITECTURE.md`
- [x] Создать `docs/TODO.md`
- [x] Инициализировать Go-модуль (`go mod init`) с версией Go 1.26
- [x] Создать структуру папок по Clean Architecture:
  - [x] `cmd/server/`
  - [x] `internal/domain/`
  - [x] `internal/usecase/` + `internal/usecase/repository/`
  - [x] `internal/interfaces/http/` + `internal/interfaces/http/middleware/`
  - [x] `internal/infrastructure/config/`, `internal/infrastructure/sqlite/`, `internal/infrastructure/ratelimit/`
  - [x] `web/`
- [x] Создать `.gitignore` для Go, SQLite и временных файлов
- [x] Создать `Taskfile.yaml` — канонический набор команд разработки (`task setup`, `task gen:api`, `task build`, `task run`, `task test`, `task lint`, `task docker:build`, `task docker:run`, `task ci` и др.), см. `docs/ARCHITECTURE.md` §6.7

---

## Фаза 1.5: API-контракт на TypeSpec (Design First)

- [x] Создать `package.json` и `tspconfig.yaml` в корне проекта
- [x] Установить TypeSpec-зависимости (`npm install`)
- [x] Создать структуру `api/`:
  - `api/main.tsp`
  - `api/models.tsp`
  - `api/errors.tsp`
  - `api/routes/health.tsp`
  - `api/routes/owners.tsp`
  - `api/routes/events.tsp`
  - `api/routes/slots.tsp`
  - `api/routes/bookings.tsp`
  - `api/routes/admin.tsp`
- [x] Описать все модели: Owner, Guest, Event, Booking, Slot, Settings, ErrorResponse
- [x] Описать все эндпоинты из PRD в TypeSpec
- [x] Добавить валидации (email, UUID, time, duration > 0) и enum в TypeSpec
- [x] Сгенерировать `api/generated/openapi.yaml` через `npm run build:api` (сейчас: `task gen:api`)
- [x] Обновить `docs/PRD.md`: заменить дублирующее описание API ссылкой на TypeSpec
- [x] Обновить `docs/ARCHITECTURE.md`: добавить `api/` как контракт Interface Adapters
- [x] Обновить `docs/TODO.md`: отметить выполненные задачи по TypeSpec

---

## Фаза 2: Конфигурация и доменный слой

- [x] Создать `config.yaml` с секциями `server`, `health`, `rate_limit`, `admin` (`owner_uuid`), `default` (включая `timezone` и `events`), `seed`
- [x] Реализовать пакет `internal/infrastructure/config` для чтения YAML
  - [x] Описать доменные сущности в `internal/domain/`:
  - [x] `Owner` (с `UUID`, `Timezone`, настройками календаря внутри структуры, без `SlotDuration`)
  - [x] `Guest` (отдельная сущность)
  - [x] `Event` (uuid, owner_id, name, description, duration_minutes, is_active)
  - [x] `Booking` (с `EventID`)
  - [x] `Slot`
  - [x] Доменные ошибки (`ErrNotFound`, `ErrConflict`, `ErrInvalidInput`, `ErrSlotUnavailable`, `ErrOverlap`, `ErrRateLimit`)
- [x] Реализовать утилиту генерации `uuid` (UUID v4)

---

## Фаза 3: Репозитории и миграции

- [x] Описать интерфейсы репозиториев в `internal/usecase/repository/` (порты use case-слоя):
  - [x] `OwnerRepository` (с методами `GetByID`, `GetByUUID`, `GetByEmail`, `GetAll`, `Create`, `Update`)
  - [x] `GuestRepository` (с методами `GetByID`, `GetByEmail`, `Create`)
  - [x] `EventRepository` (`GetByOwnerID`, `GetByUUID`, `GetByID`, `Create`, `Update`)
  - [x] `BookingRepository`
  - [x] `OwnerProvisioner.CreateOwnerWithDefaultEvents`, выполняющий создание владельца и его default events в одной транзакции
- [x] Реализовать SQLite-репозитории в `internal/infrastructure/sqlite/`:
  - [x] Инициализация БД и миграции (`db.go`) — `owners`, `guests`, `events`, `bookings`
  - [x] `OwnerRepository`
  - [x] `GuestRepository`
  - [x] `EventRepository`
  - [x] `BookingRepository` с транзакционной проверкой пересечений и `UNIQUE(owner_id, start_time, end_time)`
- [x] Реализовать seed-миграции с тестовым владельцем (Bob) и двумя событиями (15 и 30 минут), применив к владельцу дефолтные `OwnerSettings` из `config.yaml` (`default`)
- [x] Добавить индексы: `idx_bookings_owner_id`, `idx_bookings_event_id`, `idx_bookings_time_range`, `idx_events_owner_id`, `idx_events_uuid`, `idx_owners_uuid`, `idx_owners_email`, `idx_guests_email`

---

## Фаза 4: Бизнес-логика (Use Cases)

- [x] Реализовать `OwnerUsecase`
  - [x] Получение списка активных владельцев (без email и id)
  - [x] Получение владельца по UUID
  - [x] Создание нового владельца (явная кнопка) с валидацией email и атомарным созданием двух дефолтных событий из `default.events` (15 и 30 минут)
  - [x] Список бронирований владельца (`GET /api/owners/{uuid}/bookings`) с сортировкой по `start_time`
- [x] Реализовать `EventUsecase`
  - [x] Получение списка активных событий владельца
  - [x] Создание события с валидацией name и duration_minutes
  - [x] Проверка уникальности названия события в рамках owner
  - [x] Установка `is_active=true` по умолчанию при создании события
- [x] Реализовать `SlotUsecase`
  - [x] Проверка рабочего дня в часовом поясе owner
  - [x] Получение события по `event_uuid` и проверка его принадлежности owner
  - [x] Генерация слотов длительностью `event.duration_minutes` на 14 дней, начиная с текущей даты
  - [x] Пометка прошедших слотов текущего дня как `unavailable` с `reason: "past"`
  - [x] Пометка пересекающихся с бронями слотов как `unavailable` с `reason: "booked"`
  - [x] Возврат результата в виде `map[date][]Slot` с полями `status` и `reason` (конверт `SlotsResponse` собирается в `SlotHandler`)
- [x] Реализовать `BookingUsecase`
  - [x] Валидация входных данных (email, формат UUID v4 для `owner_uuid`/`event_uuid`, границы слота, 14-дневное окно)
  - [x] Получение события по `event_uuid` и проверка его принадлежности owner
  - [x] Проверка доступности слота в часовом поясе owner
  - [x] Проверка, что бронирование не в прошлом времени
  - [x] Создание или поиск гостя
  - [x] Создание бронирования через транзакционный репозиторий с `event_id`
  - [x] Mock-отправка email владельцу в логи
- [x] Реализовать `AdminUsecase`
  - [x] Получение владельца по UUID из config.yaml (admin.owner_uuid, default owner Bob)
  - [x] Список предстоящих бронирований default owner (start_time >= now, сортировка по start_time)
  - [x] Список всех событий default owner (активные и неактивные)
  - [x] Создание события для default owner (уникальность name, валидация duration_minutes)

---

## Фаза 4.5: Юнит-тесты бизнес-логики

Тестовые сценарии — `docs/TESTING.md` (разделы 2–6): happy path, альтернативы, ошибочные сценарии, границы. По TDD тесты пишутся до/параллельно реализации use case'ов (Фаза 4), опираясь только на сигнатуры интерфейсов из `docs/ARCHITECTURE.md` §6.

- [x] Создать in-memory fake/stub реализации репозиториев в `internal/usecase/fake/`:
  - [x] `fake.OwnerRepository`
  - [x] `fake.GuestRepository`
  - [x] `fake.EventRepository`
  - [x] `fake.BookingRepository` (с имитацией транзакционной проверки пересечений и unique constraint)
- [x] Написать тесты для `OwnerUsecase`
  - [x] Получение списка владельцев
  - [x] Получение владельца по UUID
  - [x] Создание владельца с дублирующимся email
  - [x] При создании владельца автоматически создаются два дефолтных события (15 и 30 минут)
- [x] Написать тесты для `EventUsecase`
  - [x] Получение списка событий владельца
  - [x] Создание события
  - [x] Отказ при дублирующемся названии события у одного owner
- [x] Написать тесты для `SlotUsecase`
  - [x] Генерация слотов на 14 дней, начиная с фиксированной "сегодняшней" даты (clock provider)
  - [x] Пустой результат на выходной день
  - [x] Пометка прошедших слотов текущего дня как `unavailable` с `reason: "past"`
  - [x] Пометка занятых слотов как `unavailable` с `reason: "booked"`
  - [x] Генерация слотов с разными длительностями событий (15, 30 минут)
  - [x] Работа с часовым поясом, включая отказ для несуществующего или неоднозначного local time при DST
  - [x] Возврат результата в виде `map[date][]Slot` с полями `status` и `reason`
- [x] Написать тесты для `BookingUsecase`
  - [x] Успешное создание бронирования
  - [x] Отказ при пересечении со существующей бронью
  - [x] Отказ при невалидном слоте (выходной, вне рабочих часов, не кратен длительности события)
  - [x] Отказ при бронировании с событием другого owner
  - [x] Отказ при бронировании за пределами 14-дневного окна
  - [x] Отказ при бронировании в прошлом времени
  - [x] Отказ для несуществующего или неоднозначного local time при DST
  - [x] Создание гостя при бронировании
- [x] Написать тесты для `AdminUsecase`
  - [x] Получение default owner
  - [x] Список предстоящих бронирований (включая только start_time >= now)
  - [x] Сортировка предстоящих бронирований по start_time
  - [x] Список всех событий default owner
  - [x] Создание события для default owner
  - [x] Отказ при дублирующемся названии события у default owner
- [x] Сверять наборы кейсов каждого use case с `docs/TESTING.md`:
  - [x] `OwnerUsecase` — TC-2.1…2.10
  - [x] `EventUsecase` — TC-3.1…3.9
  - [x] `SlotUsecase` — TC-4.1…4.10
  - [x] `BookingUsecase` — TC-5.1…5.19
  - [x] `AdminUsecase` — TC-6.1…6.7
- [x] Запустить тесты командой `task test` (эквивалент `go test -v -race -count=1 ./...`) и убедиться, что все проходят

---

## Фаза 5: HTTP API, middleware и раздача статики

- [x] Реализовать хендлеры в `internal/interfaces/http/`:
  - [x] `OwnerHandler` (`GET /api/owners`, `GET /api/owners/{uuid}`, `GET /api/owners/{uuid}/bookings`, `POST /api/owners`)
  - [x] `EventHandler` (`GET /api/owners/{uuid}/events`, `POST /api/owners/{uuid}/events`)
  - [x] `SlotHandler` (`GET /api/owners/{uuid}/slots?event_uuid=<event-uuid>`)
    - [x] Собрать `SlotsResponse`: `event_uuid`, `event_name`, `duration_minutes`, `timezone`, `start_date`, `end_date`, `slots`
  - [x] `BookingHandler` (`POST /api/bookings`)
  - [x] `AdminHandler` (`GET /api/admin`, `GET /api/admin/bookings`, `GET /api/admin/events`, `POST /api/admin/events`)
  - [x] `HealthHandler` (`GET /healthz`)
  - [x] `StaticHandler` (раздача `index.html` и статики)
  - [x] Валидация формата UUID v4 для path-параметра `{uuid}` во всех handler'ах
- [x] Настроить роутер (`internal/interfaces/http/router.go`) с учётом `/api/owners/{uuid}`, `/api/owners/{uuid}/events`, `/api/admin/*`
- [x] Реализовать middleware:
  - [x] Логирование
  - [x] Recover от паник
  - [x] CORS (если нужно)
  - [x] Rate limiting для `POST /api/bookings`, `POST /api/owners`, `POST /api/owners/{uuid}/events` и `POST /api/admin/events`
- [x] Реализовать сопоставление доменных ошибок с HTTP-статусами (400/404/409/429/500)
- [x] Реализовать точку входа `cmd/server/main.go` с graceful shutdown и `http.Server.Shutdown`

---

## Фаза 5.5: HTTP API тесты (black-box)

Тестовые сценарии — `docs/TESTING.md`, разделы 1–7. Реализуются как Go-тесты уровня HTTP (`httptest.Server` + роутер); статусы и тела ответов валидируются против `api/generated/openapi.yaml`.

- [x] Превратить Gherkin-кейсы `docs/TESTING.md` в исполняемые HTTP-тесты:
  - [x] §1 Healthcheck — TC-1.1
  - [x] §2 Владельцы — TC-2.1…2.10
  - [x] §3 События — TC-3.1…3.9
  - [x] §4 Слоты — TC-4.1…4.10
  - [x] §5 Бронирования — TC-5.1…5.19
  - [x] §6 Админка — TC-6.1…6.7
  - [x] §7 Rate limiting — TC-7.1…7.3
- [x] Проверить, что коды статусов и тела ошибок соответствуют `api/generated/openapi.yaml` (200/201/400/404/409/429/500, схема `ErrorResponse`)

---

## Фаза 6: Frontend (SPA на чистом HTML/JS/CSS)

- [x] Добавить `web/fonts/` — локальные шрифты Cal Sans и Inter (OFL) с `@font-face` и `font-display: swap` (по `docs/DESIGN.md`)
- [x] Создать `web/css/tokens.css` — дизайн-токены: палитра, типографика, радиусы, тени, светлая/тёмная темы (по `docs/DESIGN.md`)
- [x] Создать `web/css/styles.css` — базовые стили и компоненты: кнопки, карточки, таблицы, формы, бейджи слотов, календарная сетка, прогресс-индикатор, навигация, пустые состояния (по `docs/DESIGN.md`)
- [x] Создать `web/index.html` — базовый HTML-шаблон с подключением шрифтов и токенов
- [x] Создать `web/js/api.js` — клиент для REST API
- [x] Создать `web/js/router.js` — простой клиентский роутинг
- [x] Создать `web/js/app.js` — логика приложения
- [x] Применить темы по `docs/DESIGN.md`: тёмная (`.theme-dark`) для личного кабинета `/office`, светлая для `/`, `/owners` и визарда `/owners/{uuid}`
- [x] Реализовать экраны:
  - [x] Приветственная страница (`/`): описание MVP и ссылки на `/owners` и `/office`
  - [x] Личный кабинет владельца (`/office`): вкладка «Бронирования» и вкладка «Типы встреч» с формой создания события
  - [x] Публичный список владельцев (`/owners`)
  - [x] Персональная страница владельца (`/owners/{uuid}`): Шаг 1 (информация + события + бронирования)
  - [x] Шаг 2: выбор события (типа встречи)
  - [x] Шаг 3: календарь на 14 дней + список всех слотов в часовом поясе owner
  - [x] Шаг 3: `available` слоты выбираемы, `unavailable` слоты отображаются disabled с пометкой «забронировано» для `reason: "booked"`
  - [x] Шаг 4: форма ввода имени и email с HTML5-валидацией
  - [x] Шаг 5: success-страница с деталями брони и кнопкой «Вернуться к календарю владельца»
  - [x] Форма создания нового владельца
- [x] Frontend не отображает внутренние `id` и email владельцев/гостей

---

## Фаза 7: Docker и инфраструктура

- [x] Создать multi-stage `Dockerfile`:
  - [x] Stage 1: сборка Go-бинарника с `golang:1.26-alpine`
  - [x] Stage 2: минимальный финальный образ `alpine:3.20`
  - [x] Копирование бинарника, `config.yaml` и статики
  - [x] `HEALTHCHECK` для `/healthz`
  - [x] `mkdir -p /app/data` для SQLite
- [x] Проверить запуск одной командой `task docker:run` (сборка образа — `task docker:build`)
- [x] Убедиться, что база SQLite инициализируется и миграции применяются
- [x] Обновить `README.md` инструкциями по запуску с volume

---

## Фаза 8: Тестирование и финальная проверка

- [x] Проверить локальный запуск бэкенда (`task run`)
- [x] Проверить API через `curl`:
  - [x] `GET /healthz`
  - [x] `GET /api/owners`
  - [x] `GET /api/owners/{uuid}`
  - [x] `GET /api/owners/{uuid}/events`
  - [x] `GET /api/owners/{uuid}/slots?event_uuid=<event-uuid>`
  - [x] `GET /api/owners/{uuid}/bookings`
  - [x] `POST /api/bookings`
  - [x] `POST /api/owners/{uuid}/events`
  - [x] `POST /api/owners`
  - [x] `GET /api/admin`
  - [x] `GET /api/admin/bookings`
  - [x] `GET /api/admin/events`
  - [x] `POST /api/admin/events`
- [x] Проверить личный кабинет в браузере (`/office`): вкладка бронирований и вкладка создания события
- [x] Проверить ручное прохождение визарда в браузере (`/owners/{uuid}`)
- [x] Проверить, что после бронирования гость **не** появляется в списке владельцев
- [x] Проверить создание владельца через форму «Создать свой календарь» (`/owners/new`)
- [x] Проверить, что у нового владельца после создания автоматически есть 2 события (15 и 30 минут) через `GET /api/owners/{uuid}/events`
- [x] Проверить, что у Bob созданы 2 события (15 и 30 минут) после миграции
- [x] Проверить логи бэкенда на наличие mock-email сообщения
- [x] Проверить, что endpoint слотов возвращает 14 дней, начиная с сегодняшней даты
- [x] Проверить, что прошедшие слоты текущего дня не отдаются
- [x] Проверить rate limiting (превышение лимита → 429)
- [x] Интеграционные тесты конкурентного доступа по `docs/TESTING.md` §8:
  - [x] Параллельное бронирование одного слота — ровно один успех (TC-8.1)
  - [x] Параллельное бронирование пересекающихся слотов (TC-8.2)
  - [x] Параллельное создание владельцев с одним email (TC-8.3)
  - [x] Параллельное создание событий с одним названием (TC-8.4)
- [x] Прогнать полный набор тестов из `docs/TESTING.md` (API §1–7, конкурентный доступ §8, UI §9) и убедиться, что все проходят
- [x] Проверить сборку и запуск в Docker
- [x] Проверить graceful shutdown (`Ctrl+C`)
- [x] Пройтись по `TODO.md` и отметить все пункты выполненными

---

## Фаза 9: Документирование и сдача

- [ ] Обновить `README.md`:
  - [ ] Описание проекта
  - [ ] Стек технологий
  - [ ] Инструкции по локальному запуску
  - [ ] Инструкции по запуску в Docker с volume
  - [ ] Описание API (кратко или ссылка на `docs/PRD.md`)
  - [ ] Админская панель (default owner = первый владелец, без аутентификации)
  - [ ] События (events) и длительность встреч
  - [ ] Часовые пояса и UTC
  - [ ] Rate limiting
- [ ] Проверить, что все файлы спецификации актуальны
- [ ] Подготовить итоговый отчёт для пользователя

---

## Статус

Текущий статус: **Фаза 1 завершена**, **Фаза 1.5 (TypeSpec API-контракт) завершена**, **Фаза 2 (конфигурация и доменный слой) завершена**, **Фаза 3 (репозитории и миграции) завершена**, **Фаза 4 (бизнес-логика/use cases) завершена**, **Фаза 4.5 (юнит-тесты бизнес-логики) завершена**, **Фаза 5 (HTTP API, middleware и раздача статики) завершена**, **Фаза 5.5 (HTTP API тесты) завершена**, **Фаза 6 (Frontend, SPA на чистом HTML/JS/CSS) завершена**, **Фаза 7 (Docker и инфраструктура) завершена**, **Фаза 8 (Тестирование и финальная проверка) завершена**. Реализованы: конфигурация (`internal/infrastructure/config`), доменный слой (`internal/domain`), утилита UUID v4, SQLite-хранилище с миграциями, индексами и seed-данными (`internal/infrastructure/sqlite`), порты репозиториев (`internal/usecase/repository`), транзакционный `BookingRepository` с проверкой пересечений, use cases (`internal/usecase`): `OwnerUsecase`, `EventUsecase`, `SlotUsecase`, `BookingUsecase`, `AdminUsecase`, in-memory fake-репозитории (`internal/usecase/fake/`), юнит-тесты по `docs/TESTING.md` (TC-2.1…2.10, TC-3.1…3.9, TC-4.1…4.10, TC-5.1…5.19, TC-6.1…6.7), HTTP-хендлеры, роутер, middleware (логирование, recover, CORS, rate limiting), раздача статики, точка входа `cmd/server/main.go` с graceful shutdown, HTTP-тесты по `docs/TESTING.md` (TC-1.1, TC-2.1…2.10, TC-3.1…3.9, TC-4.1…4.10, TC-5.1…5.19, TC-6.1…6.7, TC-7.1…7.3) и фронтенд: локальные шрифты (Cal Sans, Inter), дизайн-токены и стили (`web/css/tokens.css`, `web/css/styles.css`), SPA (`web/index.html`, `web/js/api.js`, `web/js/router.js`, `web/js/app.js`) с клиентским роутингом (`/`, `/office`, `/owners`, `/owners/new`, `/owners/{uuid}`), приветственной страницей на `/`, личным кабинетом владельца на `/office` (тёмная тема), списком владельцев, формой создания владельца и 5-шаговым визардом бронирования (выбор события, календарь на 14 дней, слоты, форма гостя, success-страница) и Docker-инфраструктура: multi-stage `Dockerfile` (`golang:1.26-alpine` → `alpine:3.20` с бинарником, `config.yaml`, статикой и `tzdata`), `HEALTHCHECK` по `/healthz`, `.dockerignore`, volume `calendar-data` для SQLite (проверены инициализация БД, миграции, seed, персистентность и полный API-флоу в контейнере), инструкции по запуску с volume в `README.md`. В Фазе 8 выполнены: локальный запуск (`task run`), проверка всех эндпоинтов API через `curl` (healthz, owners, events, slots, bookings, admin — GET и POST), ручная проверка админки и визарда бронирования в браузере, проверка, что гость после бронирования **не** появляется в списке владельцев, создание владельца через форму `/owners/new` с автогенерацией 2 дефолтных событий (15 и 30 минут), проверка mock-email в логах, проверка слотов на 14 дней с исключением прошедших, проверка rate limiting (burst → 429), добавлены интеграционные HTTP-тесты конкурентного доступа `internal/interfaces/http/concurrency_test.go` по `docs/TESTING.md` §8 (TC-8.1 параллельное бронирование одного слота — ровно один успех, TC-8.2 пересекающиеся слоты, TC-8.3 дубли email, TC-8.4 дубли названий событий), полный набор тестов (`task test`: юнит §4.5 + HTTP §5.5 + конкурентность §8), `task vet`, `task lint` (0 issues), сборка и запуск в Docker (HEALTHCHECK healthy, миграции, seed, персистентность через volume, booking + mock email в контейнере), graceful shutdown по `Ctrl+C`. Следующий шаг — **Фаза 9: Документирование и сдача**.
