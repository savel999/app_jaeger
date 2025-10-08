.PHONY: up
up: down
	docker compose up -d
	go run main.go
	open "http://localhost:16686/search";


.PHONY: down
down:
	docker compose down -v
