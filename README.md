### Hexlet tests and linter status:
[![Actions Status](https://github.com/ivan-rudev/ai-for-developers-project-387/actions/workflows/hexlet-check.yml/badge.svg)](https://github.com/ivan-rudev/ai-for-developers-project-387/actions)

## Calendar MVP — запись на встречи

Сервис записи на встречи: владельцы календарей создают публичные страницы, гости
выбирают тип встречи, дату и свободный слот и бронируют время. Приветственная
страница на `/` описывает проект; личный кабинет владельца (админка, без
аутентификации) на `/office` показывает предстоящие бронирования и позволяет
управлять типами встреч.

Спецификация и детали: `docs/PRD.md`, `docs/ARCHITECTURE.md`, API-контракт — `api/generated/openapi.yaml` (TypeSpec в `api/`).

## Термины

| Термин | Описание |
|--------|----------|
| Owner | Владелец календаря: публичная страница записи и настройки доступности |
| Guest | Пользователь, который бронирует слот |
| Event | Тип встречи (название, описание, длительность) |
| Slot | Свободный временной интервал для бронирования |
| Booking | Бронирование: конкретная встреча гостя у владельца |

## Стек

- Backend: Go 1.26 (Clean Architecture), SQLite, стандартный `net/http`
- Frontend: SPA на чистом HTML/CSS/JS (без фреймворков)
- Контракт: TypeSpec → OpenAPI
- Инфраструктура: Docker (multi-stage), Taskfile

## Локальный запуск

```bash
task setup        # инструменты и npm-зависимости
task run          # сервер на http://localhost:8080
```

Порт сервера берётся из переменной окружения `PORT` (по умолчанию 8080, как в
`config.yaml`). Запустить на другом порту:

```bash
PORT=9000 task run   # сервер на http://localhost:9000
```

БД создаётся в `data/calendar.db`, применяются миграции и seed-данные
(владелец Bob + два типа встреч). Точки для проверки: `GET /healthz`,
приветственная страница `/`, личный кабинет `/office`, публичная страница
`/owners`.

## Запуск в Docker

```bash
task docker:build   # docker build -t calendar-mvp .
task docker:run     # docker run -p 8080:8080 -e PORT=8080 -v calendar-data:/app/data calendar-mvp
```

- Контейнер слушает порт из переменной `PORT` (по умолчанию 8080); `task docker:run`
  пробрасывает его наружу и передаёт в контейнер. Другой порт:

  ```bash
  PORT=9000 task docker:run   # docker run -p 9000:9000 -e PORT=9000 ...
  ```

- Личный кабинет владельца — `http://localhost:<PORT>/office`; приветственная
  страница — `http://localhost:<PORT>/`.
- SQLite-база лежит в `data/calendar.db` внутри контейнера; volume
  `calendar-data` монтируется в `/app/data` и сохраняет базу между
  перезапусками (пересоздание контейнера не стирает данные).
- Healthcheck по `GET /healthz` (wget в образе).
- Для работы с таймзонами в образ включён `tzdata`.

Удалить volume при желании начать с чистых seed-данных:

```bash
docker rm -f calendar-mvp
docker volume rm calendar-data
```

## Status and limitations

This repository currently contains the design-first specification for a booking MVP.
It is an intentionally unauthenticated educational demo: admin endpoints and guest
email addresses are not protected. Do not deploy it to a public or shared network
or use real personal data.

The planned SQLite deployment supports one application instance with one local
database volume. Horizontal scaling is out of scope.
