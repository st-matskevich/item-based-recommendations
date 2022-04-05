include .env

migrate-up:
	migrate -path migrations -database $(DATABASE_URL) up

migrate-down:
	migrate -path migrations -database $(DATABASE_URL) down

migrate-drop:
	migrate -path migrations -database $(DATABASE_URL) drop

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

run:
	go run .