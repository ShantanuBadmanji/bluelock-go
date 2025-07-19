# Variables
DB_PATH = ./database.db
MIGRATIONS_DIR = shared/database/migrations
GOOSE_CMD = goose -dir $(MIGRATIONS_DIR) sqlite3 $(DB_PATH)

# Default target
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make build          - Build all binaries"
	@echo "  make run-puller     - Run datapuller"
	@echo "  make run-authsync   - Run authsync"
	@echo ""
	@echo "Database:"
	@echo "  make db-setup       - Setup database and run initial migrations"
	@echo "  make db-up          - Run all migrations"
	@echo "  make db-down        - Rollback one migration"
	@echo "  make db-reset       - Reset all migrations"
	@echo "  make db-status      - Check migration status"
	@echo "  make db-create NAME=name - Create a new migration"
	@echo ""
	@echo "SQLC:"
	@echo "  make sqlc-generate  - Generate Go code from SQL queries"
	@echo "  make sqlc-verify    - Verify SQLC configuration"
	@echo ""
	@echo "Development:"
	@echo "  make format         - Format code"
	@echo "  make check          - Compile all Go packages"
	@echo "  make test           - Run tests"
	@echo "  make clean          - Remove binaries and clean artifacts"
	@echo "  make deps           - Install Go module dependencies"

# Build commands
.PHONY: build
build:
	go build -o bin/datapuller ./cmd/datapuller
	go build -o bin/authsync ./cmd/authsync

.PHONY: run-puller
run-puller:
	go run ./cmd/datapuller

.PHONY: run-authsync
run-authsync:
	go run ./cmd/authsync

# Database migration commands
.PHONY: db-up
db-up:
	$(GOOSE_CMD) up

.PHONY: db-down
db-down:
	$(GOOSE_CMD) down

.PHONY: db-reset
db-reset:
	$(GOOSE_CMD) reset

.PHONY: db-status
db-status:
	$(GOOSE_CMD) status

.PHONY: db-create
db-create:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make db-create NAME=your_migration_name"; \
		exit 1; \
	fi
	cd $(MIGRATIONS_DIR) && goose create $(NAME) sql

# SQLC commands
.PHONY: sqlc-generate
sqlc-generate:
	sqlc generate

.PHONY: sqlc-verify
sqlc-verify:
	sqlc verify

# Development commands
.PHONY: clean
clean:
	rm -rf bin/
	go clean

.PHONY: format
format:
	go fmt ./...

.PHONY: check
check:
	go build ./...

.PHONY: test
test:
	go test ./...

# Install dependencies
.PHONY: deps
deps:
	go mod tidy
	go mod download

# Setup database (create and run initial migrations)
.PHONY: db-setup
db-setup:
	touch $(DB_PATH)
	$(GOOSE_CMD) up