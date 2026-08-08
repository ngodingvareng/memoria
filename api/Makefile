# ==============================================================================
# VARIABLES & CONFIGURATION
# ==============================================================================
MIGRATIONS_DIR := db/migrations

# ANSI Color Codes for terminal output
COLOR_RESET  := \e[0m
COLOR_INFO   := \e[36m
COLOR_SUCCESS:= \e[32m
COLOR_WARN   := \e[33m

.PHONY: dev dev-only build test test-unit test-integration test-e2e db-generate migrate-create migrate-up migrate-down db-tables db-describe mock-generate help

# Default target: Display available commands
help:
	@printf "$(COLOR_INFO)Available Makefile commands:$(COLOR_RESET)\n"
	@printf "  $(COLOR_SUCCESS)make dev$(COLOR_RESET)              - Start Docker Compose & Go app (Auto docker-down on Ctrl+C)\n"
	@printf "  $(COLOR_SUCCESS)make dev-only$(COLOR_RESET)         - Run Go app only (without Docker)\n"
	@printf "  $(COLOR_SUCCESS)make build$(COLOR_RESET)            - Build all Go packages\n"
	@printf "  $(COLOR_SUCCESS)make test$(COLOR_RESET)             - Interactive test menu (unit, integration, e2e, all)\n"
	@printf "  $(COLOR_SUCCESS)make test-unit$(COLOR_RESET)        - Run unit tests only\n"
	@printf "  $(COLOR_SUCCESS)make test-integration$(COLOR_RESET) - Run integration tests only (repository/storage against real Postgres/RustFS)\n"
	@printf "  $(COLOR_SUCCESS)make test-e2e$(COLOR_RESET)         - Run e2e tests only (HTTP-level, full route tree against real Postgres)\n"
	@printf "  $(COLOR_SUCCESS)make migrate-create$(COLOR_RESET)   - Create a new SQL migration file\n"
	@printf "  $(COLOR_SUCCESS)make migrate-up$(COLOR_RESET)       - Apply all up migrations (credentials from .env)\n"
	@printf "  $(COLOR_SUCCESS)make migrate-down$(COLOR_RESET)     - Roll back all migrations (credentials from .env)\n"
	@printf "  $(COLOR_SUCCESS)make db-generate$(COLOR_RESET)      - Clean & regenerate DB code via sqlc\n"
	@printf "  $(COLOR_SUCCESS)make db-tables$(COLOR_RESET)        - List all tables in the database\n"
	@printf "  $(COLOR_SUCCESS)make db-describe$(COLOR_RESET)      - Describe a table's structure (prompts for name)\n"
	@printf "  $(COLOR_SUCCESS)make mock-generate$(COLOR_RESET)    - Generate mock objects using Mockery\n"
# ==============================================================================
# DEVELOPMENT
# ==============================================================================

## dev: Start Docker containers and run the Go application.
dev:
	@printf "$(COLOR_INFO)[1/3] Initializing Swagger...$(COLOR_RESET)\n"
	@$(MAKE) swagger-generate
	@printf "$(COLOR_SUCCESS)✓ Swagger initialized successfully.$(COLOR_RESET)\n"
	@printf "$(COLOR_INFO)[2/3] Starting Docker Compose...$(COLOR_RESET)\n"
	@docker-compose -f docker-compose.dev.yaml up -d
	@printf "$(COLOR_SUCCESS)✓ Docker Compose started successfully.$(COLOR_RESET)\n"
	@printf "$(COLOR_INFO)[3/3] Running Go application... (Press Ctrl+C to stop)$(COLOR_RESET)\n"
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
# SWAGGER
# ==============================================================================
swagger-generate:
	swag init -d ./cmd/api,./internal -g main.go --parseInternal --parseDependency

# ==============================================================================
# DATABASE & MIGRATE
# ==============================================================================

## migrate-create: Interactively create a new SQL migration file.
migrate-create:
	@read -p "Enter migration name: " name; \
	if [ -z "$$name" ]; then \
		printf "$(COLOR_WARN)Migration name cannot be empty!$(COLOR_RESET)\n"; \
		exit 1; \
	fi; \
	printf "$(COLOR_INFO)Creating migration file in $(MIGRATIONS_DIR)...$(COLOR_RESET)\n"; \
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $$name; \

## migrate-up: Apply all up migrations (all tables) using DB credentials from .env.
migrate-up:
	@bash -c '\
		set -a; \
		[ -f .env ] && source .env; \
		set +a; \
		missing=""; \
		for var in DATABASE_USERNAME DATABASE_PASSWORD DATABASE_DBNAME DATABASE_PORT; do \
			if [ -z "$${!var}" ]; then \
				missing="$$missing $$var"; \
			fi; \
		done; \
		if [ -n "$$missing" ]; then \
			printf "$(COLOR_WARN)Missing required env variable(s) in .env:$$missing$(COLOR_RESET)\n"; \
			exit 1; \
		fi; \
		printf "$(COLOR_INFO)Applying all up migrations...$(COLOR_RESET)\n"; \
		migrate -database "postgres://$$DATABASE_USERNAME:$$DATABASE_PASSWORD@localhost:$$DATABASE_PORT/$$DATABASE_DBNAME?sslmode=disable" -path $(MIGRATIONS_DIR) up && \
		printf "$(COLOR_SUCCESS)✓ Up migrations applied successfully.$(COLOR_RESET)\n" \
	'

## migrate-down: Roll back all migrations (all tables) using DB credentials from .env.
migrate-down:
	@bash -c '\
		set -a; \
		[ -f .env ] && source .env; \
		set +a; \
		missing=""; \
		for var in DATABASE_USERNAME DATABASE_PASSWORD DATABASE_DBNAME DATABASE_PORT; do \
			if [ -z "$${!var}" ]; then \
				missing="$$missing $$var"; \
			fi; \
		done; \
		if [ -n "$$missing" ]; then \
			printf "$(COLOR_WARN)Missing required env variable(s) in .env:$$missing$(COLOR_RESET)\n"; \
			exit 1; \
		fi; \
		printf "$(COLOR_WARN)Rolling back ALL migrations...$(COLOR_RESET)\n"; \
		migrate -database "postgres://$$DATABASE_USERNAME:$$DATABASE_PASSWORD@localhost:$$DATABASE_PORT/$$DATABASE_DBNAME?sslmode=disable" -path $(MIGRATIONS_DIR) down -all && \
		printf "$(COLOR_SUCCESS)✓ Down migrations applied successfully.$(COLOR_RESET)\n" \
	'

## db-generate: Clean existing DB files and regenerate using sqlc.
db-generate:
	@printf "$(COLOR_WARN)Cleaning up old DB files in internal/db/...$(COLOR_RESET)\n"
	@rm -f internal/db/*.go
	@printf "$(COLOR_INFO)Generating DB code using sqlc...$(COLOR_RESET)\n"
	@sqlc generate
	@printf "$(COLOR_SUCCESS)✓ DB code generated successfully.$(COLOR_RESET)\n"

## db-tables: List all tables in the database (runs psql inside the Postgres container from `make dev`).
db-tables:
	@bash -c '\
		set -a; \
		[ -f .env ] && source .env; \
		set +a; \
		missing=""; \
		for var in DATABASE_USERNAME DATABASE_DBNAME; do \
			if [ -z "$${!var}" ]; then \
				missing="$$missing $$var"; \
			fi; \
		done; \
		if [ -n "$$missing" ]; then \
			printf "$(COLOR_WARN)Missing required env variable(s) in .env:$$missing$(COLOR_RESET)\n"; \
			exit 1; \
		fi; \
		printf "$(COLOR_INFO)Listing tables in \"$$DATABASE_DBNAME\"...$(COLOR_RESET)\n"; \
		docker-compose -f docker-compose.dev.yaml exec -T postgres psql -U "$$DATABASE_USERNAME" -d "$$DATABASE_DBNAME" --pset pager=off -c "\dt" \
	'

## db-describe: Describe a table (columns, indexes, constraints); prompts for the table name.
db-describe:
	@bash -c '\
		set -a; \
		[ -f .env ] && source .env; \
		set +a; \
		missing=""; \
		for var in DATABASE_USERNAME DATABASE_DBNAME; do \
			if [ -z "$${!var}" ]; then \
				missing="$$missing $$var"; \
			fi; \
		done; \
		if [ -n "$$missing" ]; then \
			printf "$(COLOR_WARN)Missing required env variable(s) in .env:$$missing$(COLOR_RESET)\n"; \
			exit 1; \
		fi; \
		read -p "Enter table name: " table; \
		if [ -z "$$table" ]; then \
			printf "$(COLOR_WARN)Table name cannot be empty!$(COLOR_RESET)\n"; \
			exit 1; \
		fi; \
		docker-compose -f docker-compose.dev.yaml exec -T postgres psql -U "$$DATABASE_USERNAME" -d "$$DATABASE_DBNAME" --pset pager=off -c "\d $$table" \
	'

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
	@printf "  3) E2E Tests\n"
	@read -p "Enter choice [1-3]: " choice; \
	case "$$choice" in \
		1) $(MAKE) test-unit ;; \
		2) $(MAKE) test-integration ;; \
		3) $(MAKE) test-e2e ;; \
		*) printf "$(COLOR_WARN)Invalid choice. Aborting.$(COLOR_RESET)\n"; exit 1 ;; \
	esac

## test-unit: Run unit tests only.
test-unit:
	@printf "$(COLOR_INFO)Running unit tests...$(COLOR_RESET)\n"
	@go test -tags=unit ./... -v

## test-integration: Run integration tests only (repository/storage layer against real Postgres/RustFS).
test-integration:
	@printf "$(COLOR_INFO)Running integration tests...$(COLOR_RESET)\n"
	@go test -tags=integration ./... -v

## test-e2e: Run e2e tests only (HTTP-level, full route tree against real Postgres).
test-e2e:
	@printf "$(COLOR_INFO)Running e2e tests...$(COLOR_RESET)\n"
	@go test -tags=e2e ./... -v
