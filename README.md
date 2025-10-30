<!-- # Todo List

Simple todo list

## Features

- Add new tasks with a title and active date
- Mark tasks as done
- Edit existing tasks
- Delete tasks
- Filter tasks based on their completion status

## Installation & Usage

1. Clone the repository: `git clone https://github.com/erazr/todo-list.git`.
2. Navigate to the project directory: `cd todo-list`.
3. Rename .env.example to .env and change variables accordingly.
4. Start the docker containers: `make up`.
5. Navigate to swagger docs at http://localhost:8080/api/docs/index.htm.

## Libraries

1. [go-chi](https://github.com/go-chi/chi) as router
2. [zerolog](https://github.com/rs/zerolog) as logger
3. [golang-migrate](https://github.com/golang-migrate/migrate) for migrating the database -->


# Todo List Microservice

Краткое описание
---
RESTful микросервис Todo List на Go + PostgreSQL. Поддерживает создание, обновление, удаление, пометку задач как выполненные и получение списка задач по статусу. Контейнеризован с Docker, запускается через Docker Compose. Проект готов к деплою на Render.

Требования
---
- Docker
- Docker Compose
- Go 1.20+ (для локальной сборки)
- PostgreSQL (локально или через сервис)

Установка (локально)
---
1. Клонировать репозиторий:
```bash
git clone <repo-url>
cd generic-todolist
