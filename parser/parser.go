package parser

import (
	"math"

	"github.com/st-matskevich/item-based-recommendations/model"
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

func ParseUserProfile(response []model.PostTagLink) map[int]float32 {
	result := map[int]float32{}

	uniquePosts := map[int]struct{}{}
	for _, row := range response {
		//TODO: can initial value be not 0? if 0 is guranteed, if statement can be removed
		if _, ok := result[row.TagID]; !ok {
			result[row.TagID] = 1
		} else {
			result[row.TagID] += 1
		}
		uniquePosts[row.PostID] = struct{}{}
	}

	for tagID := range result {
		result[tagID] /= float32(len(uniquePosts))
	}

	normalizeVector(result)

	return result
}

func ParsePostsTags(response []model.PostTagLink) map[int]map[int]float32 {
	result := map[int]map[int]float32{}

	for _, row := range response {
		if _, ok := result[row.PostID]; !ok {
			result[row.PostID] = map[int]float32{}
		}

		result[row.PostID][row.TagID] = 1
	}

	for postID := range result {
		normalizeVector(result[postID])
	}

	return result
}
