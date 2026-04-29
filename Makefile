include .env
export

DB_URL=postgres://${DB_USER}:${DB_PASSWORD}@localhost:5432/${DB_NAME}?sslmode=disable

migrate-up:
	migrate -path db/migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path db/migrations -database "$(DB_URL)" down 1

migrate-create:
	migrate create -ext sql -dir db/migrations -seq $(name)

db-seed:
	psql "$(DB_URL)" -f db/seed.sql
