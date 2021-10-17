package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/st-matskevich/item-based-recommendations/api"
	"github.com/st-matskevich/item-based-recommendations/api/utils"
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

	id, err := strconv.ParseInt(os.Getenv("NODE_ID"), 10, 64)
	if err == nil {
		err = utils.InitSnowflakeNode(id)
	}
	if err != nil {
		log.Fatalf("Snowflake error: %v", err)
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
