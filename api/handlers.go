package api

import (
	"encoding/json"
	"net/http"

	"github.com/st-matskevich/item-based-recommendations/api/profile"
	"github.com/st-matskevich/item-based-recommendations/api/similarity"
	"github.com/st-matskevich/item-based-recommendations/db"
	"github.com/st-matskevich/item-based-recommendations/firebase"
)

const MAX_RECOMMENDED_POSTS = 5

func getRecommendationsHandler(w http.ResponseWriter, r *http.Request) HandlerResponse {
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

	topList, err := similarity.GetSimilarPosts(readers, MAX_RECOMMENDED_POSTS)
	if err != nil {
		return HandlerResponse{http.StatusInternalServerError, CreateErrorMessage("SQL_READ_ERROR"), err}
	}

	return HandlerResponse{http.StatusOK, topList, nil}
}

func getUserProfileHandler(w http.ResponseWriter, r *http.Request) HandlerResponse {
	uid, err := firebase.GetFirebaseAuth().Verify(r.Header.Get("Authorization"))
	if err != nil {
		return HandlerResponse{http.StatusBadRequest, CreateErrorMessage("AUTHORIZATION_ERROR"), err}
	}

	reader, err := profile.GetUserProfileReader(db.GetSQLClient(), uid)
	if err != nil {
		return HandlerResponse{http.StatusInternalServerError, CreateErrorMessage("SQL_QUERY_ERROR"), err}
	}

	profile, err := profile.GetUserProfile(reader)
	if err != nil {
		return HandlerResponse{http.StatusInternalServerError, CreateErrorMessage("SQL_READ_ERROR"), err}
	}

	return HandlerResponse{http.StatusOK, profile, nil}
}

func setUserProfileHandler(w http.ResponseWriter, r *http.Request) HandlerResponse {
	uid, err := firebase.GetFirebaseAuth().Verify(r.Header.Get("Authorization"))
	if err != nil {
		return HandlerResponse{http.StatusBadRequest, CreateErrorMessage("AUTHORIZATION_ERROR"), err}
	}

	input := profile.UserProfile{}
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		return HandlerResponse{http.StatusBadRequest, CreateErrorMessage("DECODER_ERROR"), err}
	}

	err = profile.SetUserProfile(db.GetSQLClient(), uid, input)
	if err != nil {
		return HandlerResponse{http.StatusInternalServerError, CreateErrorMessage("SQL_READ_ERROR"), err}
	}

	return HandlerResponse{http.StatusOK, struct{}{}, nil}
}

func corsHandler(w http.ResponseWriter, r *http.Request) HandlerResponse {
	return HandlerResponse{http.StatusOK, struct{}{}, nil}
}
