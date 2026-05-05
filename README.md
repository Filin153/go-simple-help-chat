# SimpleHelpChat

`SimpleHelpChat` — каркас backend-сервиса help/chat системы на Go.
В проекте уже выделены основные слои: `domain`, `repository`, `service`, `usecase`, `infrastructure/cache`.

## Структура проекта

### `domain`

Пакет с бизнес-моделями, DTO и общими ошибками.

- `domain/user.go`
  Пользователи, роли, manager/client DTO.
- `domain/ticket.go`
  Заявка и её статусы.
- `domain/msg.go`
  Сообщения чата, вложения, типы отправителя и статусы.
- `domain/department.go`
  Отделы и DTO для создания/обновления.
- `domain/schedule.go`
  Элементы расписания и bulk DTO для редактирования.
- `domain/refresh_token.go`
  Модель refresh-токена.
- `domain/jwt.go`
  DTO для пары JWT-токенов.
- `domain/error.go`
  Общие доменные ошибки.
- `domain/base_http_response.go`
  Базовый generic-ответ для HTTP-слоя.

`domain` не зависит от БД, HTTP и инфраструктуры.

### `internal/repository`

Базовый инфраструктурный слой для работы с PostgreSQL через `pgx`.

- `internal/repository/repository.go`
  Создаёт пул соединений, делает `SELECT 1`, открывает транзакции и даёт единый `GetDB(tx)` для работы через пул или транзакцию.

Сейчас это общий фундамент для будущих concrete-repository реализаций.

### `internal/infrastructure/cache`

Локальный in-memory cache для непрочитанных сообщений.

- `internal/infrastructure/cache/local_memory_msg_cache.go`
  Хранит сообщения по `userUUID` и `msgID`, поддерживает `Set`, `Get`, `DeleteByMsgID`.

Этот слой изолирует техническое хранение данных в памяти от usecase-логики.

### `internal/service`

Переиспользуемые технические сервисы без бизнес-сценариев.

- `internal/service/password.go`
  Хеширование и проверка паролей через `argon2id`.
- `internal/service/jwt.go`
  Создание и валидация `access` и `refresh` JWT-токенов.
- `internal/service/AES-256-GCM.go`
  Шифрование и расшифровка через AES-256-GCM.

### `internal/usecase`

Слой бизнес-сценариев.

- `internal/usecase/interface.go`
  Базовый интерфейс для открытия транзакций.
- `internal/usecase/auth.go`
  Авторизация, logout, refresh и login клиента через внешнюю систему.
- `internal/usecase/user.go`
  CRUD пользователей с валидацией длины пароля и хешированием.
- `internal/usecase/department.go`
  Создание/чтение/обновление отделов и получение расписания отдела.
- `internal/usecase/schedule.go`
  Редактирование расписания и генерация базового расписания на 365 дней.
- `internal/usecase/chat.go`
  Отправка сообщений, загрузка вложений, long-poll чтение новых сообщений, отметка сообщений как прочитанных и история чата.

#### Что делает `chat.go`

- сохраняет сообщение в транзакции;
- ограничивает вложения: максимум `10` файлов и `15MB` на файл;
- сохраняет файлы через абстракцию `S3Interface`;
- пишет metadata файлов в репозиторий;
- пушит новые сообщения в локальный cache;
- умеет читать новые сообщения через long-poll с fallback на `GetUnread`.

## Как связаны слои

Зависимости устроены так:

- `domain` не зависит ни от кого;
- `repository`, `service`, `infrastructure/cache` зависят от `domain`;
- `usecase` зависит от `domain` и интерфейсов инфраструктуры;
- `cmd` должен собирать приложение, но пока практически пуст.

Это даёт:

- изоляцию бизнес-логики от БД и внешних сервисов;
- возможность тестировать usecase через моки;
- переиспользуемые технические сервисы без копирования логики.

## Тесты и покрытие

Сейчас unit-тестами покрыты все исполняемые пакеты проекта:

- `internal/infrastructure/cache`
- `internal/repository`
- `internal/service`
- `internal/usecase`

Покрытие:

- `internal/infrastructure/cache`: `100.0%`
- `internal/repository`: `100.0%`
- `internal/service`: `100.0%`
- `internal/usecase`: `100.0%`
- `total (statements)`: `100.0%`

Проверка:

```bash
go test -count=1 ./...
go test -count=1 ./... -coverprofile=/tmp/all.cover.out
go tool cover -func=/tmp/all.cover.out
```

`cmd` и `domain` не влияют на итоговый statement coverage:

- в `cmd/main.go` пока нет рабочей логики;
- в `domain/*` в основном объявления типов, констант и ошибок.

## Что ещё не готово

В проекте всё ещё нет:

- HTTP-обработчиков и роутинга;
- concrete repository-реализаций для пользователей, тикетов, сообщений, отделов и расписания;
- интеграции с реальным S3;
- миграций БД;
- полноценной сборки приложения через `cmd/main.go`.

Итог: архитектурный каркас, основная бизнес-логика и unit-тесты уже собраны, но внешний интеграционный слой ещё предстоит реализовать.
