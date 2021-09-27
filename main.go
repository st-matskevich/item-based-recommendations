package main

import (
	"container/list"
	"fmt"
	"log"
	"net/http"

	"github.com/st-matskevich/item-based-recommendations/parser"
)

type PostSimilarity struct {
	id         int
	similarity float32
}

func getSimilarPosts(top int) *list.List {
	//mock data
	userProfileResponse := []parser.PostTagLink{
		{1, 1}, {1, 2},
		{3, 1}, {3, 3},
		{5, 1}, {5, 4},
	}

	postsTagsResponse := []parser.PostTagLink{
		{2, 1}, {2, 2},
		{4, 1}, {4, 5},
		{6, 2}, {6, 6},
		{7, 7}, {7, 8},
	}

	userProfile := parser.ParseUserProfile(userProfileResponse)
	postsTags := parser.ParsePostsTags(postsTagsResponse)
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
	return topList
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: homePage")

	//TODO: move magic "3" to consts
	topList := getSimilarPosts(3)
	for e := topList.Front(); e != nil; e = e.Next() {
		fmt.Fprintf(w, "Post %d similarity is %f\n", e.Value.(PostSimilarity).id, e.Value.(PostSimilarity).similarity)
	}
}

func handleRequests() {
	http.HandleFunc("/", homePage)
	log.Fatal(http.ListenAndServe(":10000", nil))
}

func main() {
	handleRequests()
}
