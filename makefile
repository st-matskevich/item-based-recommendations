include .env

BINARY_NAME=server.exe
BINARY_DIR=.\bin

migrate-up:
	migrate -path migrations -database $(DATABASE_URL) up

migrate-down:
	migrate -path migrations -database $(DATABASE_URL) down

migrate-drop:
	migrate -path migrations -database $(DATABASE_URL) drop

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

build:
	go build -o $(BINARY_DIR)\$(BINARY_NAME)

run: build 
	$(BINARY_DIR)\$(BINARY_NAME) 