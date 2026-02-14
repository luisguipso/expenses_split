.PHONY: dev run build test migrate-up migrate-down docker-up docker-down web-dev web-build

# Backend
dev:
	go run ./cmd/api

run: build
	./bin/api

build:
	go build -o bin/api ./cmd/api

test:
	go test ./... -v

# Database
docker-up:
	docker compose up -d db

docker-down:
	docker compose down

migrate-up:
	go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest \
		-path migrations -database "$${DATABASE_URL}" up

migrate-down:
	go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest \
		-path migrations -database "$${DATABASE_URL}" down 1

# Frontend
web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build

web-install:
	cd web && npm install

# All
up: docker-up dev
