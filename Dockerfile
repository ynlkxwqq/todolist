
FROM golang:1.23 AS builder
WORKDIR /app


ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64


RUN apt-get update && apt-get install -y gcc


COPY go.mod go.sum ./
RUN go mod download


COPY . .


RUN go build -o /todo .


FROM debian:bookworm-slim
WORKDIR /app


RUN apt-get update && apt-get install -y sqlite3


COPY --from=builder /todo .


CMD ["/app/todo"]
