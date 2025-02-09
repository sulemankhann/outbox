.PHONY: up down  test

up:
	docker-compose up -d

down:
	docker-compose down -v

test:
	go test -v ./...
