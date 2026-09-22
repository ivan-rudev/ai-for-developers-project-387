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

## API

REST API описан в `api/generated/openapi.yaml` (источник правды — TypeSpec-контракт
в `api/`). Краткая сводка эндпоинтов:

| Метод | Путь | Назначение |
|-------|------|------------|
| GET | `/healthz` | Healthcheck |
| GET | `/api/owners` | Список активных владельцев |
| GET | `/api/owners/{uuid}` | Информация о владельце и настройки доступности |
| POST | `/api/owners` | Создание владельца |
| GET | `/api/owners/{uuid}/events` | Список активных событий владельца |
| POST | `/api/owners/{uuid}/events` | Создание события |
| GET | `/api/owners/{uuid}/bookings` | Публичный список бронирований |
| GET | `/api/owners/{uuid}/slots?event_uuid=…` | Слоты на 14 дней |
| POST | `/api/bookings` | Создание бронирования |
| GET | `/api/admin` | Информация о default owner |
| GET | `/api/admin/bookings` | Предстоящие бронирования default owner |
| GET | `/api/admin/events` | События default owner |
| POST | `/api/admin/events` | Создание события от имени default owner |

Детальные примеры запросов/ответов, коды статусов и схема ошибок — в `docs/PRD.md`
(§4–5) и в TypeSpec-исходниках (`api/`).

## События и длительность встреч

**Event** — тип встречи владельца (название, описание, длительность). При создании
владельца автоматически создаются два события по умолчанию из `config.yaml`
(`default.events`): «Короткая встреча» на 15 минут и «Стандартная встреча» на 30 минут.
Название события уникально в рамках одного владельца. Дополнительные типы встреч можно
создавать в админке (`/office`) или через API (`POST /api/owners/{uuid}/events`,
`POST /api/admin/events`).

## Часовые пояса и UTC

Все времена в базе хранятся в UTC. Рабочие часы задаются настройками владельца:
`work_start`/`work_end` и рабочие дни `working_days` (по умолчанию 09:00–18:00,
Пн–Пт, часовой пояс `Europe/Moscow`). Слоты считаются и отображаются в часовом поясе
владельца; бронирования через API возвращаются в UTC (ISO 8601, суффикс `Z`).

## Админская панель

Личный кабинет владельца на `/office` — это привязка к владельцу, UUID которого указан
в `config.yaml` (`admin.owner_uuid`). В MVP это seed-владелец Bob — **первый владелец**,
создаваемый миграцией при старте. Панель работает **без аутентификации** и показывает
email гостей, поэтому сервис — учебное демо: не разворачивайте его в публичной или
общей сети. В кабинете две вкладки: предстоящие бронирования и типы встреч (с созданием
нового события).

## Rate limiting

Rate limiting (in-memory, `golang.org/x/time/rate`) применяется к публичным мутациям:
`POST /api/bookings`, `POST /api/owners`, `POST /api/owners/{uuid}/events`,
`POST /api/admin/events`. Лимит — 30 запросов/мин с одного IP при burst = 10; при
превышении возвращается `429 Too Many Requests`. Настраивается в `config.yaml`
(секция `rate_limit`).

## ИИ-агенты: OpenCode GitHub workflows

Репозиторий содержит четыре GitHub Actions workflow, запускающих агента
OpenCode (`anomalyco/opencode/github@latest`) с моделью `opencode/big-pickle`.
Все запуски требуют secret `OPENCODE_API_KEY` и используют `share: false`.
Запуски, инициированные ботами (`[bot]`), отсекаются условием `if`; исключение —
плановый аудит по расписанию (его actor — `github-actions[bot]`), который
разрешён намеренно. Ручной запуск аудита через `workflow_dispatch` по-прежнему
заблокирован для ботов.

| Workflow | Событие (триггер) | Модель | Назначение | Место результатов |
|----------|-------------------|--------|------------|-------------------|
| `opencode-review.yml` — авто-ревью PR | `pull_request`: opened, synchronize, reopened, ready_for_review | `opencode/big-pickle` | Ревью человеческих PR (PR от ботов и release-please пропускаются): поиск багов и рискованных изменений, оценка читаемости и поддерживаемости, actionable-комментарии | Комментарии в PR |
| `opencode-triage.yml` — триаж issues | `issues`: opened | `opencode/big-pickle` | Триаж новых issues от аккаунтов старше 30 дней (issues от `[bot]` пропускаются): ссылки на документацию и код, предложение подхода и рекомендаций по обработке ошибок, добавление меток. Не комментирует, если добавить нечего | Комментарий и метки на issue |
| `opencode-audit.yml` — еженедельный аудит кода | `schedule`: еженедельно в среду 05:00 UTC; ручной запуск `workflow_dispatch` (требуется промпт, модель опционально) | `opencode/big-pickle` (по умолчанию) | Поиск TODO/FIXME/HACK в исходниках и `docs/TODO.md`; отчёт пишется во временный файл вне рабочего дерева (`$RUNNER_TEMP/audit-report.md`), рабочее дерево остаётся чистым; при находках создаётся один GitHub issue (skip, если похожий уже открыт) | Опционально GitHub issue (создаётся через `gh issue create --body-file`); отчёт выкладывается артефактом `audit-report` |
| `opencode-comment.yml` — ответ на команду | `issue_comment` и `pull_request_review_comment`: created при команде `/oc` или `/opencode` | `opencode/big-pickle` | Ассистент по `code`-командам: анализ (explain/разбери/проанализируй/диагностируй) отвечает структурированным комментарием «причина → затронутые части → путь исправления» без веток и PR; запрос на реализацию (fix/исправь/создай PR и т.п.) сначала публикует тот же анализ, затем коммитит по Conventional Commits (ветку пушит и PR открывает workflow от имени OpenCode GitHub App) | Аналитический комментарий и/или PR |

### Команды

- `/oc` и `/opencode` в комментарии к issue или PR (в начале строки или после
  пробела) запускают `opencode-comment.yml`.
- Авто-ревью (`opencode-review.yml`) и триаж (`opencode-triage.yml`) запускаются
  автоматически: открытый/обновлённый PR и открытая issue соответственно.
- Аудит запускается автоматически по расписанию (среда, 05:00 UTC) или вручную:
  Actions → «Weekly Code Audit» → `Run workflow` → ввести промпт задачи.
- Прогоны и их логи смотреть во вкладке **Actions** репозитория (Actions → имя
  workflow → конкретный запуск); отчёт аудита дополнительно выложен артефактом
  на странице запуска.

### Примечания

- `share: false` во всех workflow: сессии OpenCode публично не публикуются
  (агент не создаёт share-ссылки). Это сознательное решение — агенты работают
  с приватным кодом репозитория, контентом issues и PR; публикация сессий не
  несёт пользы и расширяет поверхность утечки кода/данных.
- `concurrency` дедуплицирует запуски внутри workflow: для review — по номеру
  PR, для comment — по номеру issue/PR; одновременно могут идти не более одного
  запуска.
- Cross-workflow блокировки нет: авто-ревью на push и `/oc`-комментарий могут
  выполняться параллельно — это намеренное поведение.
- Агентные workflow не используют `continue-on-error`. `comment` пушит ветку и
  открывает PR от имени OpenCode GitHub App (OIDC-токен, `persist-credentials:
  false`), поэтому PR триггерит обычный CI; агент только коммитит по Conventional
  Commits и не выполняет `git push`/`gh pr create`. `audit` не пушит вовсе: отчёт
  пишется вне рабочего дерева и выкладывается артефактом, issue создаётся через
  `gh issue create`.

## Самооценка: первый проход и итерации

Короткая самооценка работы агента над проектом (по фактической истории
коммитов).

**С первого прохода закрыто:**

- Продуктовый MVP (фазы 1–8 `docs/TODO.md`: домен, use cases,
  SQLite-репозитории, HTTP-слой, тесты, SPA, Docker) — вошёл в первые
  коммиты и в дальнейшем не переделывался.
- Базовая CI-обвязка (build, ci, lint, test, vet, typespec-check) и
  документация (README, AGENTS.md, docs/*).

**Потребовались итерации:**

- OpenCode workflow — самая итеративная область: триаж (добавлен недостающий
  checkout), авто-ревью (skip release-please, concurrency, фикс YAML-tag,
  миграция на `OPENCODE_API_KEY`/big-pickle), аудит (write access, отчёт
  в корень репозитория, artifact, запрет пуша session-ветки, tolerate
  post-push, еженедельный cron), промпт `/oc` (analysis-first, Conventional
  Commits).
- Lighthouse — правки конфигурации: Node 20/24, environment на уровне job,
  переименование отчёта, upload target, index redirect, favicon.
- Release-please — ручной триггер релиза, skip OpenCode-ревью для release-PR.
- Единичные продуктовые фиксы: vet-ошибка `NewRouter` в тестах, лимит размера
  тела запроса.

**Вывод:** продукт собран преимущественно с первого прохода; основное число
итераций пришлось на автоматизацию CI и агентные workflow.

## Status and limitations

This repository currently contains the design-first specification for a booking MVP.
It is an intentionally unauthenticated educational demo: admin endpoints and guest
email addresses are not protected. Do not deploy it to a public or shared network
or use real personal data.

The planned SQLite deployment supports one application instance with one local
database volume. Horizontal scaling is out of scope.
