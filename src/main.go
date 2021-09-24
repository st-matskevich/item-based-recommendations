package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
)

func normalizeVector(vector map[int]float32) {
	magnitude := float32(0)
	for _, val := range vector {
		magnitude += val * val
	}
	magnitude = float32(math.Sqrt(float64(magnitude)))

	for id, val := range vector {
		val /= magnitude
		vector[id] = val
	}
}

//mock user data
func getUserProfile() map[int]float32 {
	result := map[int]float32{}

	result[1] = 1
	result[2] = 0.33333
	result[3] = 0.33333
	result[4] = 0.33333

	normalizeVector(result)

	return result
}

//mock posts data
func getPostsTags() map[int]map[int]float32 {
	result := map[int]map[int]float32{}

	result[2] = map[int]float32{}
	result[2][1] = 1
	result[2][2] = 1
	normalizeVector(result[2])

	result[4] = map[int]float32{}
	result[4][1] = 1
	result[4][5] = 1
	normalizeVector(result[4])

	result[6] = map[int]float32{}
	result[6][6] = 1
	result[6][2] = 1
	normalizeVector(result[6])

	result[7] = map[int]float32{}
	result[7][7] = 1
	result[7][8] = 1
	normalizeVector(result[7])

	return result
}

func getSimilarPosts() map[int]float32 {
	userProfile := getUserProfile()
	postsTags := getPostsTags()
	postsSimilarity := map[int]float32{}

	similarity := float32(0)

	for postID, tagsMap := range postsTags {
		//TODO: use go coroutines
		similarity = 0
		for tagID, tagWeight := range tagsMap {
			if val, ok := userProfile[tagID]; ok {
				similarity += tagWeight * val
			}
		}
		postsSimilarity[postID] = similarity
	}
	return postsSimilarity
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: homePage")

	postsSimilarity := getSimilarPosts()
	for postID, weight := range postsSimilarity {
		fmt.Fprintf(w, "Post %d similarity is %f\n", postID, weight)
	}
}

func handleRequests() {
	http.HandleFunc("/", homePage)
	log.Fatal(http.ListenAndServe(":10000", nil))
}

func main() {
	handleRequests()
}
