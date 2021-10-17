package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/st-matskevich/item-based-recommendations/api"
	"github.com/st-matskevich/item-based-recommendations/db"
	"github.com/st-matskevich/item-based-recommendations/firebase"
)

const SERVER_ADDR = ":10000"

func startRouter() {
	router := api.MakeRouter()

	log.Printf("Listening on %s", SERVER_ADDR)
	log.Fatal(http.ListenAndServe(SERVER_ADDR, router))
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	if err := db.OpenDB(os.Getenv("SQL_CONNECTION_STRING")); err != nil {
		log.Fatalf("SQL error: %v", err)
	}

	if err := firebase.OpenFirebaseClient(); err != nil {
		log.Fatalf("Firebase error: %v", err)
	}
}

func main() {
	startRouter()
}
