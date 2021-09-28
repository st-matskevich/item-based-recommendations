package main

import (
	"container/list"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/st-matskevich/item-based-recommendations/db"
	"github.com/st-matskevich/item-based-recommendations/parser"
)

type PostSimilarity struct {
	id         int
	similarity float32
}

func getSimilarPosts(user, top int) (*list.List, error) {
	response, err := db.GetUserProfile(user)
	if err != nil {
		return nil, err
	}
	userProfile, err := parser.ParseUserProfile(response)
	if err != nil {
		return nil, err
	}

	response, err = db.GetPostsTags(user)
	if err != nil {
		return nil, err
	}
	postsTags, err := parser.ParsePostsTags(response)
	if err != nil {
		return nil, err
	}

	topList := list.New()
	similarity := float32(0)

	for postID, tagsMap := range postsTags {
		//TODO: use go coroutines
		similarity = 0
		for tagID, tagWeight := range tagsMap {
			if val, ok := userProfile[tagID]; ok {
				similarity += tagWeight * val
			}
		}

		e := topList.PushBack(PostSimilarity{postID, similarity})
		for e.Prev() != nil &&
			e.Prev().Value.(PostSimilarity).similarity < e.Value.(PostSimilarity).similarity {
			topList.MoveBefore(e, e.Prev())
		}

		if topList.Len() > top {
			topList.Remove(topList.Back())
		}
	}
	return topList, nil
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: homePage")

	//TODO: move magic "1", "3" to consts
	topList, err := getSimilarPosts(1, 3)
	if err != nil {
		log.Printf("getSimilarPosts error: %v\n", err)
		return
	}

	for e := topList.Front(); e != nil; e = e.Next() {
		fmt.Fprintf(w, "Post %d similarity is %f\n", e.Value.(PostSimilarity).id, e.Value.(PostSimilarity).similarity)
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
