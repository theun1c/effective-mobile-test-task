.DEFAULT_GOAL := help

SHELL := /usr/bin/bash

-include .env

COMPOSE ?= docker compose
APP_SERVICE ?= app
DB_SERVICE ?= postgres

APP_HOST_PORT ?= 8080
APP_URL ?= http://127.0.0.1:$(APP_HOST_PORT)

DB_NAME ?= subscriptions
DB_USER ?= subscriptions

MIGRATION_UP ?= migrations/000001_create_subscriptions.up.sql
MIGRATION_DOWN ?= migrations/000001_create_subscriptions.down.sql

.PHONY: help dev-up dev-down dev-reset logs wait-postgres wait-http migrate-up migrate-down smoke-test

help:
	@printf "Available targets:\n"
	@printf "  make dev-up       Start PostgreSQL, apply migration, build and start app\n"
	@printf "  make dev-down     Stop application and database containers\n"
	@printf "  make dev-reset    Stop containers and remove volumes for a clean state\n"
	@printf "  make migrate-up   Apply subscriptions migration if it is not applied yet\n"
	@printf "  make migrate-down Roll back subscriptions migration if it exists\n"
	@printf "  make logs         Show recent application logs\n"
	@printf "  make smoke-test   Run CRUDL smoke test against the local app\n"

dev-up:
	@$(COMPOSE) up -d $(DB_SERVICE)
	@$(MAKE) wait-postgres
	@$(MAKE) migrate-up
	@$(COMPOSE) up -d --build $(APP_SERVICE)
	@$(MAKE) wait-http
	@printf "Application is ready at %s\n" "$(APP_URL)"
	@printf "Swagger UI: %s/swagger/\n" "$(APP_URL)"

dev-down:
	@$(COMPOSE) down

dev-reset:
	@$(COMPOSE) down -v --remove-orphans

logs:
	@$(COMPOSE) logs app --tail=100

wait-postgres:
	@printf "Waiting for PostgreSQL...\n"
	@until $(COMPOSE) exec -T $(DB_SERVICE) pg_isready -U "$(DB_USER)" -d "$(DB_NAME)" >/dev/null 2>&1; do \
		sleep 1; \
	done
	@printf "PostgreSQL is ready\n"

wait-http:
	@printf "Waiting for HTTP server...\n"
	@until curl -fsS "$(APP_URL)/healthz" >/dev/null 2>&1; do \
		sleep 1; \
	done
	@printf "HTTP server is ready\n"

migrate-up:
	@set -euo pipefail; \
	current_state="$$( $(COMPOSE) exec -T $(DB_SERVICE) psql -U "$(DB_USER)" -d "$(DB_NAME)" -At -c "select to_regclass('public.subscriptions');" )"; \
	if [ "$$current_state" = "subscriptions" ]; then \
		printf "Migration already applied\n"; \
	else \
		$(COMPOSE) exec -T $(DB_SERVICE) psql -U "$(DB_USER)" -d "$(DB_NAME)" -v ON_ERROR_STOP=1 < "$(MIGRATION_UP)"; \
		printf "Migration applied\n"; \
	fi

migrate-down:
	@set -euo pipefail; \
	current_state="$$( $(COMPOSE) exec -T $(DB_SERVICE) psql -U "$(DB_USER)" -d "$(DB_NAME)" -At -c "select to_regclass('public.subscriptions');" )"; \
	if [ "$$current_state" = "subscriptions" ]; then \
		$(COMPOSE) exec -T $(DB_SERVICE) psql -U "$(DB_USER)" -d "$(DB_NAME)" -v ON_ERROR_STOP=1 < "$(MIGRATION_DOWN)"; \
		printf "Migration rolled back\n"; \
	else \
		printf "Migration is not applied\n"; \
	fi

smoke-test:
	@set -euo pipefail; \
	health_status="$$(curl -sS -o /tmp/health.out -w "%{http_code}" "$(APP_URL)/healthz")"; \
	test "$$health_status" = "200"; \
	post_status="$$(curl -sS -o /tmp/subscription-post.json -w "%{http_code}" -X POST "$(APP_URL)/subscriptions" -H "Content-Type: application/json" -d '{"service_name":"SmokeTestPlus","price":400,"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"07-2025"}')"; \
	test "$$post_status" = "201"; \
	subscription_id="$$(python3 -c 'import json; print(json.load(open("/tmp/subscription-post.json"))["id"])')"; \
	get_status="$$(curl -sS -o /tmp/subscription-get.json -w "%{http_code}" "$(APP_URL)/subscriptions/$$subscription_id")"; \
	test "$$get_status" = "200"; \
	list_status="$$(curl -sS -o /tmp/subscription-list.json -w "%{http_code}" "$(APP_URL)/subscriptions")"; \
	test "$$list_status" = "200"; \
	python3 -c 'import json,sys; data=json.load(open("/tmp/subscription-list.json")); target=sys.argv[1]; assert any(item["id"] == target for item in data["subscriptions"])' "$$subscription_id"; \
	put_status="$$(curl -sS -o /tmp/subscription-put.json -w "%{http_code}" -X PUT "$(APP_URL)/subscriptions/$$subscription_id" -H "Content-Type: application/json" -d '{"service_name":"SmokeTestPlusUpdated","price":500,"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"07-2025","end_date":"12-2025"}')"; \
	test "$$put_status" = "200"; \
	delete_status="$$(curl -sS -o /tmp/subscription-delete.out -w "%{http_code}" -X DELETE "$(APP_URL)/subscriptions/$$subscription_id")"; \
	test "$$delete_status" = "204"; \
	missing_status="$$(curl -sS -o /tmp/subscription-missing.json -w "%{http_code}" "$(APP_URL)/subscriptions/$$subscription_id")"; \
	test "$$missing_status" = "404"; \
	swagger_status="$$(curl -sS -o /tmp/swagger.yaml -w "%{http_code}" "$(APP_URL)/swagger/openapi.yaml")"; \
	test "$$swagger_status" = "200"; \
	grep -q "openapi: 3.0.3" /tmp/swagger.yaml; \
	printf "Smoke test passed for %s\n" "$(APP_URL)"
