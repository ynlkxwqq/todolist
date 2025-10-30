# --- Build stage ---
FROM golang:1.23 AS builder
WORKDIR /app

# Включаем CGO (обязательно для sqlite3)
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

# Устанавливаем компилятор C (для sqlite3)
RUN apt-get update && apt-get install -y gcc

# Копируем зависимости и скачиваем модули
COPY go.mod go.sum ./
RUN go mod download

# Копируем остальной код
COPY . .

# Собираем бинарник
RUN go build -o /todo .

# --- Final stage ---
FROM debian:bookworm-slim
WORKDIR /app

# Устанавливаем SQLite, если нужно
RUN apt-get update && apt-get install -y sqlite3

# Копируем бинарник
COPY --from=builder /todo .

# Запускаем сервер
CMD ["/app/todo"]
