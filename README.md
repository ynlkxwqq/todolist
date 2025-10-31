# Todo List Microservice
## Project Overview
Todo List is a RESTful microservice for managing personal tasks.
It allows users to create, update, delete, and mark tasks as completed.
The service also supports filtering by task status (active / done) and automatically adds the prefix
“WEEKEND - ” to tasks scheduled on Saturdays or Sundays.
## Technologies Used
Go (Golang) — main programming language
SQLite — lightweight embedded database
Docker — containerization
Docker Compose — service orchestration
Render — deployment and hosting platform
## Project Structure
.
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
├── go.sum
├── main.go
├── internal/
│   └── ... (business logic, handlers, configuration)
└── README.md
## Requirements
You’ll need the following tools installed:
Go 1.22+
Docker
Docker Compose
Render Account (for deployment)
## Installation & Run
 1. Clone the repository
git clone https://github.com/<your-username>/todo-list.git
cd todo-list
 2. Build and start using Docker Compose
docker-compose up --build
The server will be available at:
http://localhost:8080
 3. Using Makefile (optional)
You can also manage the project with Makefile:
make build   # build the binary
make run     # run locally
make up      # start via docker-compose
make down    # stop containers
#� Deploying to Render
Go to Render.com and create a Web Service.
Connect your GitHub repository.
Set Build Command:
docker build -t todo-list .
Set Start Command:
./todo
Render will automatically build and deploy your service.
## API Endpoints
️ Create a new task
POST /api/todo-list/tasks
{
  "title": "Buy a book",
  "activeAt": "2025-10-30"
}
 Response:
{
  "id": "46861dbd-3ac1-4fad-bca0-849771302688"
}
Status codes
201 Created — task created successfully
404 Not Found — task with same title and activeAt already exists
400 Bad Request — invalid data
 Update an existing task
PUT /api/todo-list/tasks/{ID}
{
  "title": "Buy a book - updated",
  "activeAt": "2025-10-31"
}
 Response: 204 No Content
Errors:
404 Not Found — task not found
 Delete a task
DELETE /api/todo-list/tasks/{ID}
 Response: 204 No Content
Errors:
404 Not Found — task not found
 Mark a task as completed
PUT /api/todo-list/tasks/{ID}/done
 Response: 204 No Content
Errors:
404 Not Found — task not found
 Get all tasks
GET /api/todo-list/tasks?status=active
or
GET /api/todo-list/tasks?status=done
# Example response:
[
  {
    "id": "65f19340848f4be025160391",
    "title": "Buy a book - High Performance Applications",
    "activeAt": "2023-08-05"
  },
  {
    "id": "75f19340848f4be025160392",
    "title": "Buy an apartment",
    "activeAt": "2023-08-05"
  },
  {
    "id": "45f19340848f4be025160394",
    "title": "Buy a car",
    "activeAt": "2023-08-05"
  }
]
Notes:
If the date is a weekend, the task title will automatically include "WEEKEND - ".
Returns an empty array [] if no tasks exist.
Response code: 200 OK.
# Data Validation
Field	Required	Rule
title	✅	Must be ≤ 200 characters
activeAt	✅	Must be a valid date in YYYY-MM-DD format
Uniqueness	✅	title + activeAt must be unique
🧾 Curl Test Examples
Create a task:
curl -X POST https://todolist-4-p567.onrender.com/api/todo-list/tasks \
-H "Content-Type: application/json" \
-d '{"title":"Buy a book","activeAt":"2025-10-30"}'
Get all tasks:
curl https://todolist-4-p567.onrender.com/api/todo-list/tasks
Delete a task:
curl -X DELETE https://todolist-4-p567.onrender.com/api/todo-list/tasks/46861dbd-3ac1-4fad-bca0-849771302688


server: https://todolist-4-p567.onrender.com or https://todolist-4-p567.onrender.com/api/todo-list/tasks
