.PHONY: build test test-race test-cover test-integration lint fmt vet tidy run clean up down ps logs migrate-up migrate-down migrate-status migrate-new

# Local stack (Postgres, Kafka, Redis, Kafka UI). `up` waits until all healthchecks pass.
up:
	docker compose up -d --wait

down:
	docker compose down

ps:
	docker compose ps

logs:
	docker compose logs -f --tail=100

migrate-up:
	$(GOOSE) up

migrate-down:
	$(GOOSE) down

migrate-status:
	$(GOOSE) status

# Usage: make migrate-new NAME=add_something
migrate-new:
	@if [ -z "$(NAME)" ]; then echo "usage: make migrate-new NAME=<name>"; exit 1; fi
	go run github.com/pressly/goose/v3/cmd/goose@v3.29.0 -dir migrations create $(NAME) sql

# Usage: make run SERVICE=service-registry
SERVICE ?=

# Database migrations (goose, pinned and run on demand, no go.mod dependency).
# Defaults match docker-compose.yml; override DATABASE_URL for other databases.
DATABASE_URL ?= postgres://monitoring:monitoring@localhost:5432/monitoring?sslmode=disable
GOOSE := go run github.com/pressly/goose/v3/cmd/goose@v3.29.0 -dir migrations postgres "$(DATABASE_URL)"

build:
	go build ./...

run:
	@if [ -z "$(SERVICE)" ]; then echo "usage: make run SERVICE=<name>"; exit 1; fi
	go run ./cmd/$(SERVICE)

test:
	go test ./...

test-race:
	go test -race ./...

# Needs Docker: testcontainers starts a throwaway Postgres per package.
test-integration:
	go test -race -tags=integration -count=1 ./...

test-cover:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1

lint: vet
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed; run 'brew install golangci-lint'. Falling back to go vet only."; \
	fi

fmt:
	gofmt -l -w .

vet:
	go vet ./...

tidy:
	go mod tidy

clean:
	rm -f coverage.out
	go clean ./...
