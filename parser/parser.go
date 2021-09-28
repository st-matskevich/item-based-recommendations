package parser

import (
	"math"

	"github.com/st-matskevich/item-based-recommendations/db/model"
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

func ParseUserProfile(fetcher model.PostTagLinkFetcher) (map[int]float32, error) {
	result := map[int]float32{}

	uniquePosts := map[int]struct{}{}
	row := model.PostTagLink{}

	ok, err := fetcher.Next(&row)
	for ; ok; ok, err = fetcher.Next(&row) {
		//TODO: can initial value be not 0? if 0 is guranteed, if statement can be removed
		if _, contains := result[row.TagID]; !contains {
			result[row.TagID] = 1
		} else {
			result[row.TagID] += 1
		}
		uniquePosts[row.PostID] = struct{}{}
	}

	if err != nil {
		return nil, err
	}

	for tagID := range result {
		result[tagID] /= float32(len(uniquePosts))
	}

	normalizeVector(result)

	return result, nil
}

func ParsePostsTags(fetcher model.PostTagLinkFetcher) (map[int]map[int]float32, error) {
	result := map[int]map[int]float32{}

	row := model.PostTagLink{}
	ok, err := fetcher.Next(&row)
	for ; ok; ok, err = fetcher.Next(&row) {
		if _, contains := result[row.PostID]; !contains {
			result[row.PostID] = map[int]float32{}
		}

		result[row.PostID][row.TagID] = 1
	}

	if err != nil {
		return nil, err
	}

	for postID := range result {
		normalizeVector(result[postID])
	}

	return result, nil
}
