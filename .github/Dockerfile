FROM golang:1.21.1-alpine3.14 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/bin/ ./cmd/...
