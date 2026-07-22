# ==============================================================================
# VARIABLES & CONFIGURATION
# ==============================================================================
MIGRATIONS_DIR := db/migrations

# ANSI Color Codes for terminal output
COLOR_RESET  := \e[0m
COLOR_INFO   := \e[36m
COLOR_SUCCESS:= \e[32m
COLOR_WARN   := \e[33m

.PHONY: dev dev-only build test test-unit test-integration db-generate migrate-create mock-generate help

# Default target: Display available commands
help:
	@printf "$(COLOR_INFO)Available Makefile commands:$(COLOR_RESET)\n"
	@printf "  $(COLOR_SUCCESS)make dev$(COLOR_RESET)              - Start Docker Compose & Go app (Auto docker-down on Ctrl+C)\n"
	@printf "  $(COLOR_SUCCESS)make dev-only$(COLOR_RESET)         - Run Go app only (without Docker)\n"
	@printf "  $(COLOR_SUCCESS)make build$(COLOR_RESET)            - Build all Go packages\n"
	@printf "  $(COLOR_SUCCESS)make test$(COLOR_RESET)             - Interactive test menu (unit, integration, all)\n"
	@printf "  $(COLOR_SUCCESS)make test-unit$(COLOR_RESET)        - Run unit tests only\n"
	@printf "  $(COLOR_SUCCESS)make test-integration$(COLOR_RESET) - Run integration tests only\n"
	@printf "  $(COLOR_SUCCESS)make db-generate$(COLOR_RESET)      - Clean & regenerate DB code via sqlc\n"
	@printf "  $(COLOR_SUCCESS)make migrate-create$(COLOR_RESET)   - Create a new SQL migration file\n"
	@printf "  $(COLOR_SUCCESS)make mock-generate$(COLOR_RESET)    - Generate mock objects using Mockery\n"
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

## build: Compile all packages in the project.
build:
	@printf "$(COLOR_INFO)Building application...$(COLOR_RESET)\n"
	@go build ./...
	@printf "$(COLOR_SUCCESS)✓ Build completed successfully.$(COLOR_RESET)\n"

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

## test: Interactive prompt to select test type.
test:
	@printf "$(COLOR_INFO)Select the test suite to run:$(COLOR_RESET)\n"
	@printf "  1) Unit Tests\n"
	@printf "  2) Integration Tests\n"
	@read -p "Enter choice [1-3]: " choice; \
	case "$$choice" in \
		1) $(MAKE) test-unit ;; \
		2) $(MAKE) test-integration ;; \
		*) printf "$(COLOR_WARN)Invalid choice. Aborting.$(COLOR_RESET)\n"; exit 1 ;; \
	esac

## test-unit: Run unit tests only.
test-unit:
	@printf "$(COLOR_INFO)Running unit tests...$(COLOR_RESET)\n"
	@go test ./internal/usecase -v

## test-integration: Run integration tests only.
test-integration:
	@printf "$(COLOR_INFO)Running integration tests...$(COLOR_RESET)\n"
	@go test -tags=integration ./internal/repository -v