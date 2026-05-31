include .env
export

DB_URL=postgres://${DB_USER}:${DB_PASSWORD}@localhost:5432/${DB_NAME}?sslmode=disable
VERSION ?= dev
BINARY_NAME = ct-api
BUILD_LOC = ./bin/$(BINARY_NAME)

.DEFAULT_GOAL := build

.PHONY: migrate-up
migrate-up:
	migrate -path db/migrations -database "$(DB_URL)" up

.PHONY: migrate-down
migrate-down:
	migrate -path db/migrations -database "$(DB_URL)" down 1

.PHONY: migrate-create
migrate-create:
	migrate create -ext sql -dir db/migrations -seq $(name)

.PHONY: db-seed
db-seed:
	psql "$(DB_URL)" -f db/seed.sql

.PHONY: fix
fix:
	gofmt -s -w .
	goimports -w .
	actionlint .github/workflows/*.yml
	golangci-lint run --fix

.PHONY: build
build:
	go build -trimpath -ldflags "-X main.version=$(VERSION) -w -s" -o $(BUILD_LOC) ./cmd/api	

.PHONY: lint-migrations
lint-migrations:
	squawk db/migrations/*.sql

.PHONY: test
test:
	go test -v ./internal/...
