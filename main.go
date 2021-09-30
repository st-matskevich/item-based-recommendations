package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/st-matskevich/item-based-recommendations/api/similarity"
	"github.com/st-matskevich/item-based-recommendations/db"
)

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: homePage")

	//TODO: move magic "1", "3" to consts
	topList, err := similarity.GetSimilarPosts(db.GetSQLClient(), 1, 3)
	if err != nil {
		log.Printf("GetSimilarPosts error: %v\n", err)
		return
	}

	for e := topList.Front(); e != nil; e = e.Next() {
		fmt.Fprintf(w, "Post %d similarity is %f\n", e.Value.(similarity.PostSimilarity).Id, e.Value.(similarity.PostSimilarity).Similarity)
	}
}

func handleRequests() {
	http.HandleFunc("/", homePage)
	log.Fatal(http.ListenAndServe(":10000", nil))
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	if err := db.OpenDB(os.Getenv("SQL_CONNECTION_STRING")); err != nil {
		log.Fatal(err)
	}
}

func main() {
	handleRequests()
}
