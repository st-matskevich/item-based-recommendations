include .env

migrate-up:
	migrate -source file://migrations -database $(SQL_CONNECTION_STRING) up

migrate-drop:
	migrate -source file://migrations -database $(SQL_CONNECTION_STRING) drop

run:
	go run .