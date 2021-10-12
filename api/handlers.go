package api

import (
	"net/http"

	"github.com/st-matskevich/item-based-recommendations/api/similarity"
	"github.com/st-matskevich/item-based-recommendations/db"
	"github.com/st-matskevich/item-based-recommendations/firebase"
)

const MAX_RECOMMENDED_POSTS = 5

func similarityRequest(w http.ResponseWriter, r *http.Request) HandlerResponse {
	uid, err := firebase.GetFirebaseAuth().Verify(r.Header.Get("Authorization"))
	if err != nil {
		return HandlerResponse{http.StatusBadRequest, CreateErrorMessage("AUTHORIZATION_ERROR"), err}
	}

	profileReader, err := similarity.GetUserProfileReader(db.GetSQLClient(), uid)
	if err != nil {
		return HandlerResponse{http.StatusInternalServerError, CreateErrorMessage("SQL_QUERY_ERROR"), err}
	}

	postsReader, err := similarity.GetPostsTagsReader(db.GetSQLClient(), uid)
	if err != nil {
		return HandlerResponse{http.StatusInternalServerError, CreateErrorMessage("SQL_QUERY_ERROR"), err}
	}

	readers := similarity.ProfilesReaders{
		UserProfileReader: profileReader,
		PostsTagsReader:   postsReader,
	}

	topList, err := similarity.GetSimilarPosts(&readers, MAX_RECOMMENDED_POSTS)
	if err != nil {
		return HandlerResponse{http.StatusInternalServerError, CreateErrorMessage("SQL_READ_ERROR"), err}
	}

	return HandlerResponse{http.StatusOK, topList, nil}
}
