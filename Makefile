.PHONY: help build run start stop restart clean logs ps

help: ## Show this help message
	@echo "Dataset CLI - Makefile Commands"
	@echo "================================"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Usage:"
	@echo "  make start       - Start PostgreSQL and CLI container"
	@echo "  make run        - Run CLI in interactive mode"
	@echo "  make logs       - View CLI container logs"
	@echo "  make ps         - Show running containers"
	@echo "  make stop       - Stop all containers"
	@echo "  make clean      - Remove containers and data"

build: ## Build the Docker image
	docker compose build

start: ## Start PostgreSQL database
	docker compose up -d postgres
	@echo "PostgreSQL is starting..."
	@echo "Wait for it to be healthy, then run: make run"

run: ## Run CLI in interactive mode
	docker compose run --rm dataset-cli

stop: ## Stop all containers
	docker compose stop

restart: ## Restart all containers
	docker compose restart

logs: ## View CLI container logs
	docker compose logs -f dataset-cli

ps: ## Show running containers
	docker compose ps

clean: ## Remove containers and volumes
	docker compose down -v
	@echo "All containers and data have been removed."

down: ## Stop and remove containers
	docker compose down

exec-postgres: ## Execute commands in PostgreSQL container
	docker compose exec postgres psql -U postgres -d dataset
