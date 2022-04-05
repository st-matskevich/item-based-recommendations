package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/st-matskevich/item-based-recommendations/internal/api"
	"github.com/st-matskevich/item-based-recommendations/internal/db"
	"github.com/st-matskevich/item-based-recommendations/internal/firebase"
)

func startRouter() {
	router := api.MakeRouter()
	port := os.Getenv("PORT")
	SERVER_ADDR := ":" + port
	log.Printf("Listening on %s", SERVER_ADDR)
	log.Fatal(http.ListenAndServe(SERVER_ADDR, router))
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	if err := db.OpenDB(os.Getenv("DATABASE_URL")); err != nil {
		log.Fatalf("SQL error: %v", err)
	}

	if err := firebase.OpenFirebaseClient(); err != nil {
		log.Fatalf("Firebase error: %v", err)
	}
}

func main() {
	startRouter()
}
