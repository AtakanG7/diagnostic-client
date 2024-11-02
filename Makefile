.PHONY: build run test clean

# Build settings
BINARY_NAME=diagnostic-client
BUILD_DIR=bin

build:
	@echo "Building diagnostic client API..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) cmd/api/main.go

run:
	@go run cmd/api/main.go

test:
	@go test -v ./...

clean:
	@rm -rf $(BUILD_DIR)

# Database initialization
init-db:
	@psql -h localhost -U postgres -f internal/db/schema.sql

# Docker helpers
docker-postgres:
	@docker run -d --name diagnostic-postgres \
		-e POSTGRES_HOST_AUTH_METHOD=trust \
		-p 5432:5432 \
		postgres:latest

.PHONY: init-db docker-postgres
