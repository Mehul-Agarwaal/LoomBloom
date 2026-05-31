APP_NAME := loombloom
BIN_DIR := bin
SERVER := ./cmd/server
MIGRATIONS_DIR := migrations

.PHONY: dev build run test fmt vet generate migrate-up seed-dev migrate-down

dev:
	go run $(SERVER)

generate:
	go run github.com/a-h/templ/cmd/templ generate

build:
	go build -o $(BIN_DIR)/$(APP_NAME) $(SERVER)

run: build
	$(BIN_DIR)/$(APP_NAME)

test:
	go test ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

migrate-up:
	psql "$$DATABASE_URL" -f $(MIGRATIONS_DIR)/001_init.up.sql

seed-dev:
	psql "$$DATABASE_URL" -f $(MIGRATIONS_DIR)/002_seed.dev.sql

migrate-down:
	psql "$$DATABASE_URL" -f $(MIGRATIONS_DIR)/001_init.down.sql
