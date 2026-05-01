# effective-mobile-test-task

REST-сервис для управления онлайн-подписками пользователей.

Сейчас в проекте есть реализация:

- `POST /subscriptions`
- `GET /subscriptions`
- `GET /subscriptions/{id}`
- `PUT /subscriptions/{id}`
- `DELETE /subscriptions/{id}`
- `GET /subscriptions/total`

## Требования

- Docker и Docker Compose
- `make` для рекомендуемого сценария запуска
- `curl` для ручной проверки API
- `python3` для `make smoke-test` - необязательно если нет нужды в тестах

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

Рекомендуемый способ для локальной проверки:

```bash
make dev-up
```

Эта команда:

- поднимает PostgreSQL;
- ждёт готовности БД;
- применяет миграцию, если таблица `subscriptions` ещё не создана;
- собирает и запускает приложение;
- ждёт `healthz`.

Полезные команды:

```bash
make help
make logs
make dev-down
make dev-reset
make smoke-test
```

Если нужен запуск вручную без `Makefile`:

```bash
docker compose up -d postgres
docker compose exec -T postgres \
  psql -U subscriptions -d subscriptions -v ON_ERROR_STOP=1 \
  < migrations/000001_create_subscriptions.up.sql
docker compose up -d --build app
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

Через `Makefile`:

```bash
make migrate-up
make migrate-down
```

`make dev-up` уже включает `make migrate-up`.

Если нужен ручной запуск в Docker Compose окружении:

Применить основную миграцию:

```bash
docker compose exec -T postgres \
  psql -U subscriptions -d subscriptions -v ON_ERROR_STOP=1 \
  < migrations/000001_create_subscriptions.up.sql
```

Если база уже была инициализирована раньше, повторный запуск этой команды вернёт ошибку `relation "subscriptions" already exists`. В таком случае:

- либо не применяйте миграцию повторно;
- либо используйте `make migrate-up`, который пропускает уже применённую миграцию;
- либо используйте `make dev-reset`, если нужна полностью чистая база.

## Swagger

Доступно два endpoint'а:

- Swagger docs page: `http://127.0.0.1:8080/swagger/`
- Raw OpenAPI spec: `http://127.0.0.1:8080/swagger/openapi.yaml`

Страница `/swagger/` self-contained и не требует CDN для открытия.
Оба endpoint'а используют текущий файл [docs/swagger/swagger.yaml](docs/swagger/swagger.yaml).

## CRUDL Smoke Scenario

Ниже минимальный сценарий ручной проверки `feat-001-subs`.

Автоматизированный вариант:

```bash
make smoke-test
```

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
