.PHONY: up down build run-leader run-replica  test

up:
	docker-compose up -d

down:
	docker-compose down -v

build:
	go build -o bin/example ./cmd/example

run-leader:
	go run ./cmd/example --leader --port 8080

run-replica:
	go run ./cmd/example --port 8081

test:
	go test -v ./...
