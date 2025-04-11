.PHONY: build run test migrate-up migrate-down docker-build docker-up docker-down test-db-up test-db-down test-integration test-all

# Build the application
build:
	go build -o bin/chatops ./cmd/chatops

# Run the application
run: build
	./bin/chatops

# Run tests
test:
	go test -v ./...

# Database migrations
migrate-up:
	migrate -path migrations -database "postgresql://chatops:chatops@localhost:5432/chatops?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgresql://chatops:chatops@localhost:5432/chatops?sslmode=disable" down

# Docker commands
docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

# Install dependencies
deps:
	go mod tidy
	go mod download

# Test database
test-db-up:
	docker-compose -f docker-compose.test.yml up -d

test-db-down:
	docker-compose -f docker-compose.test.yml down

# Run integration tests
test-integration: test-db-up
	go test -v ./internal/tests/integration/...
	make test-db-down

# Run all tests
test-all: test test-integration 