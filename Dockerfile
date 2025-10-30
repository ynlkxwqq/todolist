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
FROM golang:1.22.6 AS build

WORKDIR /app

ENV CGO_ENABLED=0

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /todo .

# --- Final stage ---
FROM alpine:3.18

RUN apk add --no-cache ca-certificates

COPY --from=build /todo /todo

EXPOSE 8080

ENTRYPOINT ["/todo"]
