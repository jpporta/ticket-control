ifneq (,$(wildcard ./.env))
    include .env
    export
endif

tidy:
	@echo "Running go mod tidy..."
	go mod tidy

cli:
	@echo "Running the CLI application..."
	go run ./cmd/cli/* --name="Andiara Porta"
run:
	@echo "Running the application..."
	go run ./cmd/web/*
run-printer:
	@echo "Running the printer..."
	go run ./cmd/printer/*
run-task:
	@echo "Running the task..."
	go run ./cmd/task/*

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

typst-task:
	@echo "Watch Typst task..."
	typst watch ./internal/printer/models/task.typ task.pdf

