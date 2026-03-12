# SimpleHelpChat

`SimpleHelpChat` — заготовка backend-сервиса для help/chat системы на Go.
Сейчас в проекте уже выделены основные слои: `domain`, `repository`, `service`, `usecase`.

## Что уже есть

### `domain`

Пакет с бизнес-моделями и общими типами.

- `domain/user.go`
  Описывает пользователей и связанные сущности:
  `User`, `CreateUser`, `UpdateUser`, `Manager`, `Client`, `CreateClient`.
- `domain/ticket.go`
  Описывает заявку (`Ticket`) и её статусы.
- `domain/msg.go`
  Описывает сообщения в чате, вложения и DTO для создания сообщений.
- `domain/refresh_token.go`
  Модель refresh-токена, который хранится в БД.
- `domain/jwt.go`
  DTO для пары JWT-токенов (`access_token`, `refresh_token`).
- `domain/schedule.go`
  Модели расписания и базовая валидация времени.
- `domain/department.go`
  Модель отдела.
- `domain/error.go`
  Общие доменные ошибки.
- `domain/base_http_response.go`
  Базовый формат ответа для HTTP-слоя.

Зачем нужен `domain`:
это единое место, где лежат структуры данных и бизнес-типы, без привязки к БД, HTTP и конкретным сервисам.

### `internal/repository`

Сейчас реализован базовый инфраструктурный слой для работы с PostgreSQL через `pgx`.

- `internal/repository/repository.go`
  Создаёт пул соединений, проверяет подключение, открывает транзакции и даёт единый `GetDB(tx)` для работы либо через транзакцию, либо напрямую через пул.

Зачем нужен `repository`:
он изолирует доступ к БД от бизнес-логики. В usecase-слое используются интерфейсы, а не конкретная реализация БД.

### `internal/service`

Слой прикладных сервисов без бизнес-сценариев.

- `internal/service/password.go`
  Хеширование и проверка паролей через `argon2id`.
- `internal/service/jwt.go`
  Создание и проверка `access` и `refresh` JWT-токенов.
- `internal/service/AES-256-GCM.go`
  Симметричное шифрование и расшифровка данных через AES-256-GCM.

Зачем нужен `service`:
сюда вынесены технические операции, которые можно переиспользовать в разных usecase без дублирования логики.

### `internal/usecase`

Слой бизнес-сценариев.

- `internal/usecase/interface.go`
  Базовый интерфейс для открытия транзакций.
- `internal/usecase/auth.go`
  Сценарии авторизации:
  `Login`, `LoginClient`, `Logout`, `Refresh`, `GetAccessTokenClaims`.

  Что делает:
  работает с пользователями, refresh-токенами, паролями и JWT;
  сам не знает ничего про конкретную БД и зависит только от интерфейсов.

- `internal/usecase/user.go`
  CRUD для пользователей:
  `GetAll`, `GetByUUID`, `GetByLogin`, `Create`, `UpdateByUUID`, `DeleteByUUID`.

  Что делает:
  управляет пользователями через `UserRepo`;
  при создании и обновлении хеширует пароль перед сохранением.

Зачем нужен `usecase`:
это место, где собирается бизнес-логика из domain-моделей, repository-интерфейсов и сервисов.

### `cmd`

- `cmd/main.go`
  Пока это пустая точка входа. HTTP/API слой ещё не собран.

## Как связаны слои

Схема зависимостей сейчас такая:

`domain` <- `service`

`domain` <- `repository`

`domain` + `service` + `repository interfaces` <- `usecase`

Это нужно, чтобы:

- бизнес-логика не зависела от конкретной БД;
- криптография, JWT и хеширование паролей не смешивались с usecase;
- модели данных оставались простыми и переиспользуемыми.

## Что покрыто тестами

Сейчас unit-тестами покрыты:

- `internal/service`
- `internal/usecase`

На текущий момент покрытие этих пакетов доведено до `100%`.

Проверка:

```bash
go test ./...
go test ./internal/service -cover
go test ./internal/usecase -cover
```

## Что ещё не готово

Сейчас в проекте ещё нет:

- HTTP-обработчиков / роутинга
- полноценной реализации repository-методов для пользователей, тикетов, сообщений и refresh-токенов
- миграций БД
- конфигурации приложения и запуска через `cmd/main.go`

То есть архитектурный каркас и ключевая бизнес-логика уже есть, но интеграционный слой вокруг них ещё предстоит собрать.
