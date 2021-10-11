package api

import (
	"encoding/json"
	"net/http"

	"github.com/st-matskevich/item-based-recommendations/api/similarity"
	"github.com/st-matskevich/item-based-recommendations/db"
	"github.com/st-matskevich/item-based-recommendations/firebase"
)

const MAX_RECOMMENDED_POSTS = 5

func similarityRequest(w http.ResponseWriter, r *http.Request) HandlerResponse {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	uid, err := firebase.GetFirebaseAuth().Verify(r.Header.Get("Authorization"))
	if err != nil {
		return HandlerResponse{400, err}
	}

	profileReader, err := similarity.GetUserProfileReader(db.GetSQLClient(), uid)
	if err != nil {
		return HandlerResponse{500, err}
	}

	postsReader, err := similarity.GetPostsTagsReader(db.GetSQLClient(), uid)
	if err != nil {
		return HandlerResponse{500, err}
	}

	readers := similarity.ProfilesReaders{
		UserProfileReader: profileReader,
		PostsTagsReader:   postsReader,
	}

	topList, err := similarity.GetSimilarPosts(&readers, MAX_RECOMMENDED_POSTS)
	if err != nil {
		return HandlerResponse{500, err}
	}

	err = json.NewEncoder(w).Encode(topList)
	if err != nil {
		return HandlerResponse{500, err}
	}
	return HandlerResponse{200, nil}
}
