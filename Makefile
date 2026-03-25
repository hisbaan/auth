.PHONY: help dev dev/api dev/web build build/api build/web run run/api run/web clean clean/api clean/web test lint fmt db-generate db-migrate db-status db-inspect db-validate db-hash jet

help:
	@echo "Available commands:"
	@echo "  make dev                Run all services in development mode (via overmind)"
	@echo "  make dev/api            Run only API in dev mode"
	@echo "  make dev/web            Run only web in dev mode"
	@echo "  make build              Build all services"
	@echo "  make build/api          Build the API binary"
	@echo "  make build/web          Build the web app bundle"
	@echo "  make run                Run built services via overmind"
	@echo "  make run/api            Run the built API binary"
	@echo "  make run/web            Run the built web bundle"
	@echo "  make clean              Clean build artifacts"
	@echo "  make clean/api          Clean API build artifacts"
	@echo "  make clean/web          Clean web build artifacts"

dev:
	overmind start -f Procfile

dev/api:
	$(MAKE) -C api dev

dev/web:
	pnpm --dir web dev

build: build/api build/web

build/api:
	$(MAKE) -C api build

build/web:
	pnpm --dir web build

run: # build
	overmind start -f Procfile.run

run/api: # build/api
	$(MAKE) -C api run

run/web: # build/web
	pnpm --dir web start

clean: clean/api clean/web

clean/api:
	$(MAKE) -C api clean

clean/web:
	rm -rf web/.next

test:
	@echo "No configured tests..."
