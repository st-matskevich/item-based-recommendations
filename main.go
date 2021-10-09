package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/st-matskevich/item-based-recommendations/api/similarity"
	"github.com/st-matskevich/item-based-recommendations/db"
)

const MAX_RECOMMENDED_POSTS = 5

func similarityRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	vars := mux.Vars(r)
	userId, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("GetSimilarPosts error: %v\n", err)
		w.WriteHeader(400)
		return
	}

	topList, err := similarity.GetSimilarPosts(db.GetSQLClient(), userId, MAX_RECOMMENDED_POSTS)
	if err != nil {
		log.Printf("GetSimilarPosts error: %v\n", err)
		w.WriteHeader(500)
		return
	}

	err = json.NewEncoder(w).Encode(topList)
	if err != nil {
		log.Printf("GetSimilarPosts error: %v\n", err)
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
}

func handleRequests() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/recommendations/{id}", similarityRequest)
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
	handleRequests()
}
