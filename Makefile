.PHONY: build up down migrate logs

build:
	docker build -t todo-app .

up:
	docker-compose up -d --build

down:
	docker-compose down

migrate:
	@echo "Applying SQL migrations..."
	@docker-compose exec -T db psql -U postgres -d todo -f /migrations/000001_create_tasks.up.sql

logs:
	docker-compose logs -f app
