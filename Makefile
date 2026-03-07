.PHONY: help dev dev/api dev/web build run clean test lint fmt db-generate db-migrate db-status db-inspect db-validate db-hash jet

help:
	@echo "Available commands:"
	@echo "  make dev                Run all services in development mode (via overmind)"
	@echo "  make dev/api            Run only API in dev mode"
	@echo "  make dev/web            Run only web in dev mode"
	@echo "  make build              Build all services"
	@echo "  make run                Run the API server"
	@echo "  make clean              Clean build artifacts"

dev:
	overmind start -f Procfile

build: api/build

run: api/run

clean: api/clean

test:
	@echo "Running tests..."
