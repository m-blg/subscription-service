# Subscription Service

REST service for managing user online subscriptions

## Build and Run

### Run with Docker
1. Make sure Docker and Docker Compose are installed.
2. Start the app and database:
	```shell
	make docker-up
	```
	Or directly:
	```shell
	sudo docker compose -f docker-compose.yaml up --build
	```
3. The service will be available at http://localhost:8080.

To stop containers:
	```
	make docker-down
	```
	Or:
	```
	sudo docker compose -f docker-compose.yaml down
	```

### Local Development
1. Install Go 1.25+
2. Install dependencies:
	```shell
	make deps
	```
3. (Optional) Build the binary:
	```shell
	make build
	```
	The binary will be created in `bin/subscription-service`.
5. Start db container:
	```
	sudo docker compose -f docker-compose.yaml up db --build
	```

4. Run the app:
	```shell
	make run
	```
	Or run the built binary:
	```shell
	./bin/subscription-service
	```
5. The service will be available at http://localhost:8080.

To stop db container:
	```
	make docker-down
	```
	Or:
	```
	sudo docker compose -f docker-compose.yaml db down
	```


## Swagger

http://localhost:8080/swagger/index.html

To rebuild swagger docs:
```shell
make swag
```

## Migrations

By default migrations applied during service initialization.

### Manual migration

To migrate one step up:
```shell
make migrate-up
```

To migrate one step down:
```shell
make migrate-down
```

# Testing

To run unit tests:
```shell
make test
```

