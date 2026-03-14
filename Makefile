APP_NAME=subscription-service
MAIN_PATH=cmd/api/main.go
DOCKER_COMPOSE=docker-compose.yaml

.PHONY: run build test migrate-up migrate-down swag docker-up docker-down lint deps

run:
	go run $(MAIN_PATH)

build:
	go build -o bin/$(APP_NAME) $(MAIN_PATH)

test:
	go test -v ./internal/...

migrate-up:
	$(GOPATH)/bin/goose up-by-one

migrate-down:
	$(GOPATH)/bin/goose down

swag:
	$(GOPATH)/bin/swag init -g $(MAIN_PATH)

docker-up:
	sudo docker compose -f $(DOCKER_COMPOSE) up --build

docker-down:
	sudo docker compose -f $(DOCKER_COMPOSE) down

lint:
	golangci-lint run

deps:
	go mod tidy
	go mod download
