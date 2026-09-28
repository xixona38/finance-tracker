include .env
export

.PHONY: run db-up db-down db-logs migrate-up migrate-down

run:
	@go run cmd/app/main.go

db-up:
	@docker compose up -d db

db-down:
	@docker compose stop db

db-logs:
	@docker compose logs -f db

migrate-up:
	@migrate -path ./migrations -database "$(PG_URL)" -verbose up

migrate-down:
	@migrate -path ./migrations -database "$(PG_URL)" -verbose down 1