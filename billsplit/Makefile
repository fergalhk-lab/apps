.PHONY: build test docker up down

build:
	cd frontend && npm ci && npm run build
	go build -o ../bin/billsplit .
	go build -o ../bin/seed ./cmd/seed/

test:
	go test ./...

docker:
	cd .. && docker build -t billsplit -f billsplit/Dockerfile .

up:
	docker compose up --build -d
	@echo "Run 'docker compose logs seed' to retrieve the admin invite code."

down:
	docker compose down --remove-orphans --volumes
