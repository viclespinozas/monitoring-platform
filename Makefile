.PHONY: build test test-race test-cover lint fmt vet tidy run clean up down ps logs

# Local stack (Postgres, Kafka, Redis, Kafka UI). `up` waits until all healthchecks pass.
up:
	docker compose up -d --wait

down:
	docker compose down

ps:
	docker compose ps

logs:
	docker compose logs -f --tail=100

# Usage: make run SERVICE=service-registry
SERVICE ?=

build:
	go build ./...

run:
	@if [ -z "$(SERVICE)" ]; then echo "usage: make run SERVICE=<name>"; exit 1; fi
	go run ./cmd/$(SERVICE)

test:
	go test ./...

test-race:
	go test -race ./...

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
