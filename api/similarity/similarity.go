package similarity

import (
	"math"

	"github.com/st-matskevich/item-based-recommendations/db"
)

//TODO: similarity field should be private in production
type PostSimilarity struct {
	Id         int     `json:"id"`
	Similarity float32 `json:"similarity"`
}

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

func readUserProfile(reader db.PostTagLinkReader) (map[int]float32, error) {
	result := map[int]float32{}

	uniquePosts := map[int]struct{}{}
	row := db.PostTagLink{}

	ok, err := reader.Next(&row)
	for ; ok; ok, err = reader.Next(&row) {
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

func readPostsTags(reader db.PostTagLinkReader) (map[int]map[int]float32, error) {
	result := map[int]map[int]float32{}

	row := db.PostTagLink{}
	ok, err := reader.Next(&row)
	for ; ok; ok, err = reader.Next(&row) {
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

func GetSimilarPosts(fetcher db.ProfilesFetcher, userId string, topSize int) ([]PostSimilarity, error) {
	response, err := fetcher.GetUserProfile(userId)
	if err != nil {
		return nil, err
	}
	userProfile, err := readUserProfile(response)
	if err != nil {
		return nil, err
	}

	response, err = fetcher.GetPostsTags(userId)
	if err != nil {
		return nil, err
	}
	postsTags, err := readPostsTags(response)
	if err != nil {
		return nil, err
	}

	result := []PostSimilarity{}
	similarity := float32(0)

	for postID, tagsMap := range postsTags {
		//TODO: use go coroutines
		similarity = 0
		for tagID, tagWeight := range tagsMap {
			if val, ok := userProfile[tagID]; ok {
				similarity += tagWeight * val
			}
		}

		if len(result) < topSize {
			result = append(result, PostSimilarity{postID, similarity})
		} else if similarity > result[len(result)-1].Similarity {
			result[len(result)-1] = PostSimilarity{postID, similarity}
		} else {
			continue
		}

		for idx := len(result) - 1; idx > 0 && result[idx].Similarity > result[idx-1].Similarity; idx-- {
			result[idx], result[idx-1] = result[idx-1], result[idx]
		}
	}
	return result, nil
}
