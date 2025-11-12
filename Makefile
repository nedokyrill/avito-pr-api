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
	@migrate create -ext sql -dir db/migrations -seq ${name}

migrate-up:
	@migrate -database ${POSTGRESQL_URL} -path db/migrations up

migrate-down:
	@migrate -database ${POSTGRESQL_URL} -path db/migrations down 1

