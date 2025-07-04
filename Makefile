include .env
.PHONY: run build up down logs migrateup migratedown

# Run the application locally
run:
	go run ./cmd/api

# Build the Go binary
build:
	go build -o bin/manga-api ./cmd/api

# Start the docker-compose stack in the background
up:
	docker-compose up --build -d

# Stop and remove the docker-compose stack
down:
	docker-compose down

# View logs from the api service
logs:
	docker-compose logs -f api

# Run database migrations
migrateup:
	migrate -path migrations -database "${DB_URL}" -verbose up

migratedown:
	migrate -path migrations -database "${DB_URL}" -verbose down

test:
	docker-compose up -d test_db && \
	echo "Waiting for test database to be ready..." && \
	sleep 5 && \
	go test -v -cover ./... && \
	docker-compose stop test_db

swag:
	swag init -g cmd/api/main.go
