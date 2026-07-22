# Run application
dev:
	docker-compose -f docker-compose.dev.yaml up -d && \
	go run cmd/api/main.go

dev-only:
	go run cmd/api/main.go

db-generate:
	rm -f internal/db/*.go
	sqlc generate

MIGRATIONS_DIR = db/migrations

migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $$name

mock-generate:
	mockery