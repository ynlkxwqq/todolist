# FROM golang:1.22.5-alpine AS builder

# WORKDIR /build

# COPY . .

# RUN CGO_ENABLED=0 GOOS=linux go build -o todo-list .

# FROM alpine AS hoster

# WORKDIR /app

# COPY --from=builder /build/.env ./.env
# COPY --from=builder /build/migrations ./migrations
# COPY --from=builder /build/todo-list ./todo-list

# ENTRYPOINT [ "./todo-list" ]


# Dockerfile
# --- Build stage ---
# --- Build stage ---
FROM golang:1.22.6 AS build

# Рабочая директория
WORKDIR /app

# Отключаем cgo для статической сборки
ENV CGO_ENABLED=0

# Копируем go.mod и go.sum для кеширования зависимостей
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем весь проект
COPY . .

# Собираем бинарник
RUN go build -o /todo main.go

# --- Final stage ---
FROM alpine:3.18

# Устанавливаем сертификаты для HTTPS
RUN apk add --no-cache ca-certificates

# Копируем бинарник из build stage
COPY --from=build /todo /todo

# Порт приложения
EXPOSE 8080

# Запуск бинарника
ENTRYPOINT ["/todo"]
