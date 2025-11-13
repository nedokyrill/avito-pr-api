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
	@docker compose down

# ТЕСТЫ

tests:
	@go test -cover ./...