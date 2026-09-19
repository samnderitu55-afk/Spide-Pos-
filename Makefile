.PHONY: help run build dev clean test build-linux deploy-local

help:
	@echo "Commands:"
	@echo "  make run          - Run the app locally (go run)"
	@echo "  make dev          - Live reload with air"
	@echo "  make build        - Build Windows binary (spide-pos.exe)"
	@echo "  make build-linux  - Build Linux binary (spide-pos)"
	@echo "  make test         - Run go tests"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make check        - go vet + gofmt check"

run:
	go run cmd/server/main.go

dev:
	@command -v air >/dev/null 2>&1 || (echo "Install air: go install github.com/air-verse/air@latest" && exit 1)
	air

build:
	go build -o spide-pos.exe cmd/server/main.go

build-linux:
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o spide-pos cmd/server/main.go
	@echo "Binary: spide-pos ($$(stat -c %s spide-pos) bytes)"

test:
	go test ./...

clean:
	rm -f spide-pos spide-pos.exe
	rm -rf tmp/

check:
	@echo "Running go vet..."
	go vet ./...
	@echo "Checking gofmt..."
	@test -z "$$(gofmt -l .)" || (echo "gofmt found issues:" && gofmt -l . && exit 1)
	@echo "✓ All checks passed"