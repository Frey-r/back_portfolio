.PHONY: dev build test clean init-db help

APP_NAME = portfolio-backend
BIN_DIR = bin

dev: ## Run the development server
	@echo "Starting development server..."
	@go run ./cmd/server

build: ## Build the application
	@echo "Building binary to $(BIN_DIR)/$(APP_NAME)..."
	@go build -o $(BIN_DIR)/$(APP_NAME) ./cmd/server

test: ## Run unit tests
	@echo "Running tests..."
	@go test -v ./...

clean: ## Remove built binaries
	@echo "Cleaning up..."
	@rm -rf $(BIN_DIR)

init-db: ## Initialize the data directory for SQLite
	@echo "Ensuring data directory exists..."
	@mkdir -p data

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
