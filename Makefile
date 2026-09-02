# Cabinet Studio API — slice 0
.PHONY: up down logs test run build tidy proto

up:
	docker compose up --build -d

down:
	docker compose down

logs:
	docker compose logs -f api

test:
	go test ./...

run:
	go run ./cmd/api

build:
	mkdir -p bin
	go build -o bin/api ./cmd/api

tidy:
	go mod tidy

proto:
	protoc --go_out=. --go_opt=module=github.com/Danish-13-Khan/backend-cabinet proto/cabinet/v1/outbox.proto
