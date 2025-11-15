.PHONY: generate-api clean-api build-app run new-migrate migrate-up migrate-down docker-up docker-down tests lint
-include .env

# ГЕНЕРАЦИЯ КОДА ИЗ api/openapi.yml

generate-api: clean-api
	@oapi-codegen -generate types -package generated -o internal/generated/sdk.go api/openapi.yml

clean-api:
	@rm -rf internal/generated/*.go

# ЛОКАЛЬНЫЙ ЗАПУСК ПРИЛОЖЕНИЯ

build-app:
	@go build -o ./.bin/app ./cmd/main.go

run:build-app
	@./.bin/app

# СОЗДАНИЕ И ЛОКАЛЬНЫЙ ЗАПУСК МИГРАЦИЙ

new-migrate:
	@migrate create -ext sql -dir db/migrations ${name}

migrate-up:
	@migrate -database ${DB_URL} -path db/migrations up

migrate-down:
	@migrate -database ${DB_URL} -path db/migrations down 1

# ЗАПУСК В КОНТЕЙНЕРАХ, ЧЕРЕЗ DOCKER COMPOSE

docker-up:
	@docker compose up -d --build

docker-down:
	@docker compose down -v

# ТЕСТЫ

tests:
	@go test -cover ./...

integration-tests:
	@docker compose down
	@docker compose up -d postgres migrate
	@go test -tags=integration -v ./integration_tests/...
	@docker compose down

# ЛИНТИНГ

lint:
	@golangci-lint run

