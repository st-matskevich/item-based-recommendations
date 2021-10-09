package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/st-matskevich/item-based-recommendations/api/similarity"
	"github.com/st-matskevich/item-based-recommendations/db"
)

const MAX_RECOMMENDED_POSTS = 5

func similarityRequest(w http.ResponseWriter, r *http.Request) HandlerResponse {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	vars := mux.Vars(r)
	userId, err := strconv.Atoi(vars["id"])
	if err != nil {
		return HandlerResponse{400, err}
	}

	topList, err := similarity.GetSimilarPosts(db.GetSQLClient(), userId, MAX_RECOMMENDED_POSTS)
	if err != nil {
		return HandlerResponse{500, err}
	}

	err = json.NewEncoder(w).Encode(topList)
	if err != nil {
		return HandlerResponse{500, err}
	}
	return HandlerResponse{200, nil}
}
