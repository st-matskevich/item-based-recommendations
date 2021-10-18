include .env

migrate-up:
	migrate -source file://migrations -database $(SQL_CONNECTION_STRING) up

migrate-down:
	migrate -source file://migrations -database $(SQL_CONNECTION_STRING) down

migrate-drop:
	migrate -source file://migrations -database $(SQL_CONNECTION_STRING) drop

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

run:
	go run .