.PHONY: swagger swagger-serve build run migrate test test-verbose lint

swagger:
	@echo "Generating Swagger documentation..."
	@swag init -g cmd/api/main.go -o docs/swagger

swagger-serve: swagger
	@echo "Swagger docs available at http://localhost:8081/v1/swagger/index.html"

build:
	@go build -o bin/api cmd/api/main.go
	@go build -o bin/migrate cmd/migrate/main.go

run:
	@go run cmd/api/main.go

migrate:
	@go run cmd/migrate/main.go

test:
	@go test ./...

test-verbose:
	@go test -v ./...

lint:
	@golangci-lint run ./...

