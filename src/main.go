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

func getUserProfile() map[int]float32 {
	result := map[int]float32{}

	//mock data
	response := []struct {
		postID int
		tagID  int
	}{
		{1, 1}, {1, 2},
		{3, 1}, {3, 3},
		{5, 1}, {5, 4},
	}

	uniquePosts := map[int]struct{}{}
	for _, row := range response {
		//TODO: can initial value be not 0? if 0 is guranteed, if statement can be removed
		if _, ok := result[row.tagID]; !ok {
			result[row.tagID] = 1
		} else {
			result[row.tagID] += 1
		}
		uniquePosts[row.postID] = struct{}{}
	}

	for tagID := range result {
		result[tagID] /= float32(len(uniquePosts))
	}

	normalizeVector(result)

	return result
}

func getPostsTags() map[int]map[int]float32 {
	result := map[int]map[int]float32{}

	//mock data
	response := []struct {
		postID int
		tagID  int
	}{
		{2, 1}, {2, 2},
		{4, 1}, {4, 5},
		{6, 2}, {6, 6},
		{7, 7}, {7, 8},
	}

	for _, row := range response {
		if _, ok := result[row.postID]; !ok {
			result[row.postID] = map[int]float32{}
		}

		result[row.postID][row.tagID] = 1
	}

	for postID := range result {
		normalizeVector(result[postID])
	}

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
