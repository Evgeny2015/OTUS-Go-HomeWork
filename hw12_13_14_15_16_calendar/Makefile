BIN_CALENDAR := "./bin/calendar"
BIN_SCHEDULER := "./bin/calendar_scheduler"
BIN_SENDER := "./bin/calendar_sender"
BIN_MIGRATE := "./bin/migrate"
DOCKER_IMG="calendar:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

#Postrges
POSTGRES_USER ?= postgres
POSTGRES_PASSWORD ?= password
POSTGRES_DB ?= backend
POSTGRES_PORT ?= 5435
POSTGRES_CONTAINER := postgres-calendar

#Rabbit
RABBIT_USER ?= rabbit
RABBIT_PASSWORD ?= password
RABBIT_PORT ?= 5672
RABBIT_PORT_UI ?= 15672
RABBIT_CONTAINER := rabbitmq


run-postgres:
	docker run -d --name $(POSTGRES_CONTAINER) \
	-e POSTGRES_USER=$(POSTGRES_USER) \
	-e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
	-e POSTGRES_DB=$(POSTGRES_DB) \
	-p $(POSTGRES_PORT):5432 \
	-v postgres-data:/var/lib/postgresql/data \
	postgres:latest

run-rabbit:
	docker run -d --name $(RABBIT_CONTAINER) \
	-p $(RABBIT_PORT):5672 \
	-p $(RABBIT_PORT_UI):15672 \
	-e RABBITMQ_DEFAULT_USER=$(RABBIT_USER) \
	-e RABBITMQ_DEFAULT_PASS=$(RABBIT_PASSWORD) \
	rabbitmq:management

build:
	go build -v -o $(BIN_CALENDAR) -ldflags "$(LDFLAGS)" ./cmd/calendar
	go build -v -o $(BIN_SCHEDULER) -ldflags "$(LDFLAGS)" ./cmd/calendar_scheduler
	go build -v -o $(BIN_SENDER) -ldflags "$(LDFLAGS)" ./cmd/calendar_sender
	go build -v -o $(BIN_MIGRATE) -ldflags "$(LDFLAGS)" ./cmd/migrate

run: build
	$(BIN_CALENDAR) -config ./configs/calendar_config.yml

# Migrations
migrate-up: build
	$(BIN_MIGRATE) -command up -dsn "host=localhost port=5432 user=postgres password=postgres dbname=calendar sslmode=disable"

migrate-down: build
	$(BIN_MIGRATE) -command down -dsn "host=localhost port=5432 user=postgres password=postgres dbname=calendar sslmode=disable"

migrate-status: build
	$(BIN_MIGRATE) -command status -dsn "host=localhost port=5432 user=postgres password=postgres dbname=calendar sslmode=disable"

migrate-create: build
	$(BIN_MIGRATE) -command create -dir ./migrations $(filter-out $@,$(MAKECMDGOALS))

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

version: build
	$(BIN_CALENDAR) version

test:
	go test -race ./internal/... ./pkg/...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v2.6.2

lint: install-lint-deps
	golangci-lint run ./...

generate-proto:
	protoc --go_out=. --go-grpc_out=. -I. internal/server/protobuf/*.proto

# Docker Compose
COMPOSE_FILE := deployments/docker-compose.yaml
COMPOSE_INTEGRATION_FILE := deployments/docker-compose.integration.yaml

up:
	docker-compose -f $(COMPOSE_FILE) up

down:
	docker-compose -f $(COMPOSE_FILE) down

up-detached:
	docker-compose -f $(COMPOSE_FILE) up -d

logs:
	docker-compose -f $(COMPOSE_FILE) logs -f

integration-tests:
	docker-compose -f $(COMPOSE_INTEGRATION_FILE) up --abort-on-container-exit --exit-code-from integration-tests; \
	EXIT_CODE=$$?; \
	docker-compose -f $(COMPOSE_INTEGRATION_FILE) down; \
	exit $$EXIT_CODE

.PHONY: build run build-img run-img version test lint generate-proto up down up-detached logs migrate-up migrate-down migrate-status migrate-create integration-tests
