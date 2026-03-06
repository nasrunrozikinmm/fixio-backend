.PHONY: run dev build test swagger clean

# Run the application
run:
	go run cmd/main.go

# Run with hot-reload (requires air: go install github.com/air-verse/air@latest)
dev:
	air

# Build binary
build:
	go build -o bin/fixio cmd/main.go

# Run tests
test:
	go test ./... -v

# Run tests with coverage
test-cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# Generate Swagger docs
swagger:
	swag init -g cmd/main.go -o docs --parseDependency --parseInternal

# Tidy dependencies
tidy:
	go mod tidy

# Clean build artifacts
clean:
	rm -rf bin/ docs/ coverage.out coverage.html

# Seed default data
seed:
	go run cmd/seeder/main.go
