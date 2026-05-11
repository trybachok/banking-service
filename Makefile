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
