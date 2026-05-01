# effective-mobile-test-task

REST-сервис для управления онлайн-подписками пользователей.

Сейчас в проекте реализована `feat-001-subs`:
- `POST /subscriptions`
- `GET /subscriptions`
- `GET /subscriptions/{id}`
- `PUT /subscriptions/{id}`
- `DELETE /subscriptions/{id}`

## Требования

- Docker и Docker Compose
- `curl` для ручной проверки API

Для локального запуска без Docker:
- Go 1.23+
- PostgreSQL 16+

## Конфигурация

Пример переменных окружения находится в [.env.example](.env.example).

Минимальный способ подготовить конфиг:

```bash
cp .env.example .env
```

Основные переменные:

- `APP_HOST_PORT` — порт приложения на хосте
- `DB_NAME`, `DB_USER`, `DB_PASSWORD` — параметры PostgreSQL
- `POSTGRES_HOST_PORT` — порт PostgreSQL на хосте
- `LOG_LEVEL` — уровень логирования

## Запуск через Docker Compose

Поднять PostgreSQL и приложение:

```bash
docker compose up -d --build
docker compose ps
```

Проверить, что приложение отвечает:

```bash
curl -i -sS http://127.0.0.1:8080/healthz
```

Ожидаемый ответ:

```http
HTTP/1.1 200 OK

{"status":"ok"}
```

## Локальный запуск без Docker

1. Поднимите PostgreSQL любым удобным способом.
2. Заполните `.env` на основе `.env.example`.
3. Запустите приложение:

```bash
go run ./cmd/api
```

## Миграции

Миграции пока применяются вручную.

Применить основную миграцию в Docker Compose окружении:

```bash
docker compose exec -T postgres \
  psql -U subscriptions -d subscriptions -v ON_ERROR_STOP=1 \
  < migrations/000001_create_subscriptions.up.sql
```

Если база уже была инициализирована раньше, повторный запуск этой команды вернёт ошибку `relation "subscriptions" already exists`. В таком случае:
- либо не применяйте миграцию повторно;
- либо поднимите окружение заново с чистым volume, если нужна полностью чистая база.

## Swagger

Доступно два endpoint'а:

- Swagger UI: `http://127.0.0.1:8080/swagger/`
- Raw OpenAPI spec: `http://127.0.0.1:8080/swagger/openapi.yaml`

Swagger UI использует текущий файл [docs/swagger/swagger.yaml](docs/swagger/swagger.yaml).

## CRUDL Smoke Scenario

Ниже минимальный сценарий ручной проверки `feat-001-subs`.

### 1. Создать подписку

```bash
curl -i -sS -X POST http://127.0.0.1:8080/subscriptions \
  -H 'Content-Type: application/json' \
  -d '{
    "service_name": "Yandex Plus",
    "price": 400,
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "07-2025"
  }'
```

Из ответа сохраните `id`.

### 2. Получить подписку по ID

```bash
curl -i -sS http://127.0.0.1:8080/subscriptions/<subscription_id>
```

### 3. Получить список подписок

```bash
curl -i -sS http://127.0.0.1:8080/subscriptions
```

### 4. Обновить подписку

```bash
curl -i -sS -X PUT http://127.0.0.1:8080/subscriptions/<subscription_id> \
  -H 'Content-Type: application/json' \
  -d '{
    "service_name": "Yandex Plus Updated",
    "price": 500,
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "07-2025",
    "end_date": "12-2025"
  }'
```

### 5. Удалить подписку

```bash
curl -i -sS -X DELETE http://127.0.0.1:8080/subscriptions/<subscription_id>
```

Ожидаемый статус: `204 No Content`.

### 6. Убедиться, что запись удалена

```bash
curl -i -sS http://127.0.0.1:8080/subscriptions/<subscription_id>
```

Ожидаемый ответ:

```http
HTTP/1.1 404 Not Found

{"error":"subscription not found"}
```

## Полезные файлы

- [ТЗ](docs/product/TECH_SPEC.md)
- [Архитектура](docs/product/ARCHITECTURE.md)
- [Контекст](docs/ai/CONTEXT.md)
- [Пошаговый flow реализации](docs/ai/features/feat-001-subs/FLOW.md)
- [Swagger spec](docs/swagger/swagger.yaml)
