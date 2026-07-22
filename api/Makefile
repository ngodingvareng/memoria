# ==============================================================================
# VARIABLES & CONFIGURATION
# ==============================================================================
MIGRATIONS_DIR := db/migrations

# ANSI Color Codes for terminal output
COLOR_RESET  := \e[0m
COLOR_INFO   := \e[36m
COLOR_SUCCESS:= \e[32m
COLOR_WARN   := \e[33m

.PHONY: dev dev-only db-generate migrate-create mock-generate help

# Default target: Display available commands
help:
	@printf "$(COLOR_INFO)Available Makefile commands:$(COLOR_RESET)\n"
	@printf "  $(COLOR_SUCCESS)make dev$(COLOR_RESET)             - Start Docker Compose & Go app (Auto docker-down on Ctrl+C)\n"
	@printf "  $(COLOR_SUCCESS)make dev-only$(COLOR_RESET)        - Run Go app only (without Docker)\n"
	@printf "  $(COLOR_SUCCESS)make db-generate$(COLOR_RESET)     - Clean & regenerate DB code via sqlc\n"
	@printf "  $(COLOR_SUCCESS)make migrate-create$(COLOR_RESET)  - Create a new SQL migration file\n"
	@printf "  $(COLOR_SUCCESS)make mock-generate$(COLOR_RESET)   - Generate mock objects using Mockery\n"

# ==============================================================================
# DEVELOPMENT
# ==============================================================================

## dev: Start Docker containers and run the Go application.
dev:
	@printf "$(COLOR_INFO)[1/2] Starting Docker Compose...$(COLOR_RESET)\n"
	@docker-compose -f docker-compose.dev.yaml up -d
	@printf "$(COLOR_SUCCESS)✓ Docker Compose started successfully.$(COLOR_RESET)\n"
	@printf "$(COLOR_INFO)[2/2] Running Go application... (Press Ctrl+C to stop)$(COLOR_RESET)\n"
	@bash -c '\
		cleanup() { \
			printf "\n$(COLOR_WARN)Stopping server & shutting down Docker containers...$(COLOR_RESET)\n"; \
			docker-compose -f docker-compose.dev.yaml down; \
			printf "$(COLOR_SUCCESS)✓ Docker Compose stopped successfully.$(COLOR_RESET)\n"; \
		}; \
		trap cleanup INT TERM; \
		go run cmd/api/main.go || true \
	'

## dev-only: Run Go application directly without Docker.
dev-only:
	@printf "$(COLOR_INFO)Running Go application...$(COLOR_RESET)\n"
	@go run cmd/api/main.go

# ==============================================================================
# DATABASE & MIGRATE
# ==============================================================================

## db-generate: Clean existing DB files and regenerate using sqlc.
db-generate:
	@printf "$(COLOR_WARN)Cleaning up old DB files in internal/db/...$(COLOR_RESET)\n"
	@rm -f internal/db/*.go
	@printf "$(COLOR_INFO)Generating DB code using sqlc...$(COLOR_RESET)\n"
	@sqlc generate
	@printf "$(COLOR_SUCCESS)✓ DB code generated successfully.$(COLOR_RESET)\n"

## migrate-create: Interactively create a new SQL migration file.
migrate-create:
	@read -p "Enter migration name: " name; \
	if [ -z "$$name" ]; then \
		printf "$(COLOR_WARN)Migration name cannot be empty!$(COLOR_RESET)\n"; \
		exit 1; \
	fi; \
	printf "$(COLOR_INFO)Creating migration file in $(MIGRATIONS_DIR)...$(COLOR_RESET)\n"; \
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $$name; \

# ==============================================================================
# TEST & MOCK
# ==============================================================================

## mock-generate: Generate mock implementations using Mockery.
mock-generate:
	@printf "$(COLOR_INFO)Generating mocks using Mockery...$(COLOR_RESET)\n"
	@mockery
	@printf "$(COLOR_SUCCESS)✓ Mocks generated successfully.$(COLOR_RESET)\n"