package similarity

import (
	"math"
	"net/http"

	"github.com/st-matskevich/item-based-recommendations/api/utils"
	"github.com/st-matskevich/item-based-recommendations/db"
	"github.com/st-matskevich/item-based-recommendations/firebase"
)

//TODO: similarity field should be private in production
type PostSimilarity struct {
	Id         int     `json:"id"`
	Similarity float32 `json:"similarity"`
}

type PostTagLink struct {
	PostID int
	TagID  int
}

type ProfilesReaders struct {
	UserProfileReader, PostsTagsReader db.ResponseReader
}

func getUserProfileReader(client *db.SQLClient, id int) (db.ResponseReader, error) {
	return client.Query("SELECT likes.post_id, tag_id FROM likes JOIN post_tag ON likes.post_id = post_tag.post_id WHERE user_id = $1", id)
}

func getPostsTagsReader(client *db.SQLClient, id int) (db.ResponseReader, error) {
	return client.Query("SELECT post_tag.post_id, tag_id FROM likes RIGHT JOIN post_tag ON likes.post_id = post_tag.post_id AND user_id = $1 WHERE user_id IS NULL", id)
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

func readUserProfile(reader db.ResponseReader) (map[int]float32, error) {
	result := map[int]float32{}

	uniquePosts := map[int]struct{}{}
	row := PostTagLink{}

	ok, err := reader.Next(&row.PostID, &row.TagID)
	for ; ok; ok, err = reader.Next(&row.PostID, &row.TagID) {
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

func readPostsTags(reader db.ResponseReader) (map[int]map[int]float32, error) {
	result := map[int]map[int]float32{}

	row := PostTagLink{}
	ok, err := reader.Next(&row.PostID, &row.TagID)
	for ; ok; ok, err = reader.Next(&row.PostID, &row.TagID) {
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

func getSimilarPosts(readers ProfilesReaders, topSize int) ([]PostSimilarity, error) {
	userProfile, err := readUserProfile(readers.UserProfileReader)
	if err != nil {
		return nil, err
	}

	postsTags, err := readPostsTags(readers.PostsTagsReader)
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

const MAX_RECOMMENDED_POSTS = 5

func GetRecommendationsHandler(w http.ResponseWriter, r *http.Request) utils.HandlerResponse {
	uid, err := firebase.GetFirebaseAuth().Verify(r.Header.Get("Authorization"))
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.AUTHORIZATION_ERROR), err)
	}

	profileReader, err := getUserProfileReader(db.GetSQLClient(), uid)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	postsReader, err := getPostsTagsReader(db.GetSQLClient(), uid)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	readers := ProfilesReaders{
		UserProfileReader: profileReader,
		PostsTagsReader:   postsReader,
	}

	topList, err := getSimilarPosts(readers, MAX_RECOMMENDED_POSTS)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	return utils.MakeHandlerResponse(http.StatusOK, topList, nil)
}
