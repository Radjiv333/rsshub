PROJECT_NAME=rsshub

DC=docker-compose

build:
	@echo "Building the project..."
	go build -o $(PROJECT_NAME) ./cmd/main.go

up:
	@echo "Starting $(PROJECT_NAME)..."
	$(DC) up --build

down:
	@echo "Stopping $(PROJECT_NAME)..."
	$(DC) down

restart: down up

nuke:
	@echo "Removing all containers, networks, and volumes..."
	$(DC) down -v

