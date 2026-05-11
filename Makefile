APP_NAME := banking-service

.PHONY: build run migrate

build:
	go build ./...

run:
	go run ./cmd/api

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down

migrate-version:
	go run ./cmd/migrate version

e2e:
	BASE_URL=http://localhost:8085 ./tests/e2e/smoke.sh

docker-final-check:
	docker compose -f deploy/docker-compose.yaml up -d postgres mailpit
	docker compose -f deploy/docker-compose.yaml run --rm migrator up
	docker compose -f deploy/docker-compose.yaml up -d --build api scheduler
	./tests/e2e/smoke.sh
