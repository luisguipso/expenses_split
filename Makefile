.PHONY: dev run build test test-integration migrate-up migrate-down docker-up docker-down web-dev web-build seed clean

# Backend
dev:
	go run ./cmd/api

run: build
	./bin/api

build:
	go build -o bin/api ./cmd/api

test:
	go test ./... -v

test-integration:
	CONTAS_TEST_DATABASE_URL="postgres://contas:contas@localhost:5432/contas_test?sslmode=disable" \
	go test ./internal/integration/ -v -count=1

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

seed:
	go run ./cmd/seed

clean:
	go run ./cmd/clean

# Frontend
web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build

web-install:
	cd web && npm install

# All
up: docker-up dev
