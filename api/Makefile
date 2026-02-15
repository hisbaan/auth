ifneq (,$(wildcard ./.env))
    include .env
    export
endif

.PHONY: jet db-generate db-migrate db-status db-hash db-inspect db-validate build run dev clean

# Generate go-jet SQL builder code from database
jet:
	@if [ -z "$(DATABASE_URL)" ]; then echo "DATABASE_URL is not set"; exit 1; fi
	jet -dsn=$(DATABASE_URL) -schema=public -path=./internal/jet

# Generate database migrations from HCL schema changes
# Usage: make db-generate name=add_users_table
db-generate:
	@if [ -z "$(name)" ]; then echo "Usage: make db-generate name=description"; exit 1; fi
	@if [ -z "$(DATABASE_URL)" ]; then echo "DATABASE_URL is not set"; exit 1; fi
	atlas migrate diff $(name) --env local

# Apply migrations to database
db-migrate:
	@if [ -z "$(DATABASE_URL)" ]; then echo "DATABASE_URL is not set"; exit 1; fi
	atlas migrate apply --env local

# Check migration status
db-status:
	@if [ -z "$(DATABASE_URL)" ]; then echo "DATABASE_URL is not set"; exit 1; fi
	atlas migrate status --env local

# Compute migration hash (run this after manual changes to migrations/)
db-hash:
	atlas migrate hash

# Inspect current database schema
db-inspect:
	@if [ -z "$(DATABASE_URL)" ]; then echo "DATABASE_URL is not set"; exit 1; fi
	atlas schema inspect --url $(DATABASE_URL)

# Validate migrations are correct
db-validate:
	@if [ -z "$(DATABASE_URL)" ]; then echo "DATABASE_URL is not set"; exit 1; fi
	atlas migrate validate --env local

build:
	go build -o bin/auth .

run:
	./bin/auth

dev:
	air

clean:
	rm -rf ./bin/*
