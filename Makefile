PROJECT_NAME=rsshub
DB_URL=postgres://postgres:changeme@localhost:5432/rsshub?sslmode=disable
DC=docker-compose

fetch:
	@export $$(grep -v '^#' .env | xargs) && ./rsshub fetch --race

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

migrate-up:
	migrate -path=./migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path=./migrations -database "$(DB_URL)" down

migrate-version:
	migrate -path=./migrations -database "$(DB_URL)" version