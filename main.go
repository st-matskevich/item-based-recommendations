package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/st-matskevich/item-based-recommendations/api"
	"github.com/st-matskevich/item-based-recommendations/db"
)

func startRouter() {
	router := api.MakeRouter()
	log.Fatal(http.ListenAndServe(":10000", router))
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	if err := db.OpenDB(os.Getenv("SQL_CONNECTION_STRING")); err != nil {
		log.Fatalf("SQL error: %v", err)
	}
}

func main() {
	startRouter()
}
