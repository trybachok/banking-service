APP_NAME := banking-service

.PHONY: build run migrate

build:
	go build ./...

run:
	go run ./cmd/api

migrate:
	@if command -v psql >/dev/null 2>&1 && [ -n "$$DATABASE_URL" ]; then \
		for f in migrations/*.sql; do echo "Applying $$f"; psql "$$DATABASE_URL" -f "$$f"; done; \
	else \
		echo "Set DATABASE_URL and ensure psql is installed to apply migrations."; \
	fi
