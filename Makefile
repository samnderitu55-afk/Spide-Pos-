.PHONY: help run build dev clean test

help:
	@echo "Commands: make run, make build, make dev, make clean"

run:
	go run cmd/server/main.go

build:
	go build -o bin/cosmetics-pos cmd/server/main.go

dev:
	@if command -v air > /dev/null; then air; else echo "Install air: go install github.com/cosmtrek/air@latest"; fi

clean:
	rm -rf bin/
	go clean

deps:
	go mod download
	go mod tidy