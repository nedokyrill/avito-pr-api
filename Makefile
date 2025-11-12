include .env
export

generate-api: clean-api
	@oapi-codegen -generate types -package generated -o internal/generated/sdk.go api/openapi.yml

clean-api:
	@rm -rf internal/generated/*.go

build-app:
	@go build -o ./.bin/app ./cmd/main.go

run:build-app
	@./.bin/app

new-migrate:
	@migrate create -ext sql -dir db/migrations ${name}

migrate-up:
	@migrate -database ${DB_URL} -path db/migrations up

migrate-down:
	@migrate -database ${DB_URL} -path db/migrations down 1

