ifneq (,$(wildcard ./.env))
    include .env
    export
endif

tidy:
	@echo "Running go mod tidy..."
	go mod tidy

run:
	@echo "Running the application..."
	go run ./cmd/web/*

new_migration:
	@echo "Creating new migration file..."
	goose create $(name) sql

up:
	@echo "Applying all up migrations..."
	goose up

down:
	@echo "Rolling back a single migrations..."
	goose down

generate:
	@echo "Generating code for queries..."
	sqlc generate

cli:
	@echo "Running the CLI application..."
	go run ./cmd/cli user create --name "$(name)"

typst-letter:
	@echo "Watch Typst letter..."
	typst watch ./internal/printer/models/letter.typ letter.pdf

typst-task:
	@echo "Watch Typst task..."
	typst watch ./internal/printer/models/task.typ task.pdf

typst-list:
	@echo "Watch Typst list..."
	typst watch ./internal/printer/models/list.typ list.pdf

test-printer:
	@echo "Running tests for the printer package..."
	go test -v ./internal/printer/... -count=1

typewriter:
	go run ./cmd/cli

build-typewriter:
	go build ./cmd/cli
