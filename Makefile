COMPOSE_FILE := deployment/localCalc/docker-compose.yml
COMPOSE_PROJECT := orderpack

.PHONY: test test-integration mocks up down build logs

test:
	go test ./...

test-integration:
	go test -tags=integration ./...

# testify mocks from interfaces (see .mockery.yaml)
mocks:
	go run github.com/vektra/mockery/v2@v2.52.2

up:
	docker compose -f $(COMPOSE_FILE) -p $(COMPOSE_PROJECT) up -d --build

down:
	docker compose -f $(COMPOSE_FILE) -p $(COMPOSE_PROJECT) down

build:
	docker compose -f $(COMPOSE_FILE) -p $(COMPOSE_PROJECT) build

logs:
	docker compose -f $(COMPOSE_FILE) -p $(COMPOSE_PROJECT) logs -f app
