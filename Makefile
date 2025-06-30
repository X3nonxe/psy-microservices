.PHONY: up down migrate logs

up:
	docker-compose up -d --build

down:
	docker-compose down --volumes

migrate:
	docker-compose run --rm migrate

logs:
	docker-compose logs -f auth-service

test:
	cd microservices/auth-service && go test -v ./...