APP_NAME=subscription-service
MAIN_PATH=cmd/api/main.go
DOCKER_COMPOSE=docker-compose.yaml

.PHONY: run build test api-test migrate-up migrate-down swag docker-up docker-down lint deps

run:
	go run $(MAIN_PATH)

build:
	go build -o bin/$(APP_NAME) $(MAIN_PATH)

test:
	go test -v ./internal/...

api-test:
	sudo docker compose -f tests/integration/docker-compose.yaml up -d
	go test -v ./tests/integration/... || (sudo docker compose -f tests/integration/docker-compose.yaml down && exit 1)
	sudo docker compose -f tests/integration/docker-compose.yaml down

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
