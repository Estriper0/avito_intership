
.PHONY: compose-up up down logs test


compose-up up:
	docker compose up --build -d

down:
	docker compose down

logs:
	docker compose logs -f