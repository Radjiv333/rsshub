PROJECT_NAME=rsshub

DC=docker-compose

fetch:
	@export $$(grep -v '^#' .env | xargs) && ./rsshub fetch

build:
	@echo "Building the project..."
	go build -o $(PROJECT_NAME) ./cmd/main.go

up:
	@echo "Starting $(PROJECT_NAME)..."
	$(DC) up --build

upd:
	@echo "Starting $(PROJECT_NAME)..."
	$(DC) up --build -d

down:
	@echo "Stopping $(PROJECT_NAME)..."
	$(DC) down

restart: down up

nuke:
	@echo "Removing all containers, networks, and volumes..."
	$(DC) down -v

