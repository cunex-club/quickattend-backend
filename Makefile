
ifneq (,$(wildcard ./.env))
	include .env
	export
endif

POSTGRES_HOST 	?= localhost
POSTGRES_PORT   ?= 5432
POSTGRES_USER 	?= postgres
POSTGRES_PASS 	?= postgres
POSTGRES_DB 		?= quickattend-db
POSTGRES_SCHEMA ?= public

DB_URL := postgres://$(POSTGRES_USER):$(POSTGRES_PASS)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable&search_path=$(POSTGRES_SCHEMA)

.PHONY: run tidy migrate

env:
	echo $(DB_URL)

run: 
	go run ./cmd/server

tidy:
	go mod tidy

test:
	go test -v ./...


compose-up:
	docker compose up -d

compose-down:
	docker compose down

migrate-up:
	@psql "postgres://$(POSTGRES_USER):$(POSTGRES_PASS)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable" -c "CREATE EXTENSION IF NOT EXISTS pg_trgm;" && \
	atlas migrate apply -u "$(DB_URL)" --dir file://tools/atlas/migrations

migrate-diff:
	@read -p "Enter migration name (no spaces): " name; \
	atlas migrate diff $$name \
		--dir file://tools/atlas/migrations \
		--to file://tools/atlas/schema.sql \
		--dev-url "docker://postgres/18-alpine/dev?search_path=public"

migrate:
	@if atlas schema inspect -u "$(DB_URL)" | diff -q - tools/atlas/schema.sql > /dev/null; then \
		echo "Schema is already up-to-date."; \
	else \
		echo "Schema is outdated. Running migrations..."; \
		$(MAKE) migrate-diff; \
		$(MAKE) migrate-up; \
	fi

# wipes everything
db-clean:
	@echo "Wiping everything from db..."
	atlas schema clean -u "$(DB_URL)" --auto-approve

db-seed:
	@echo "Seeding database..."
	psql "postgres://$(POSTGRES_USER):$(POSTGRES_PASS)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable" \
		-f tools/seed.sql

# reset back to seed data
db-reset: db-clean migrate-up db-seed
	@echo "Database restored to seed state."
