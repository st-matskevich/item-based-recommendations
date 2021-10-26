package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/st-matskevich/item-based-recommendations/internal/api/profile"
	"github.com/st-matskevich/item-based-recommendations/internal/api/replies"
	"github.com/st-matskevich/item-based-recommendations/internal/api/similarity"
	"github.com/st-matskevich/item-based-recommendations/internal/api/tasks"
	"github.com/st-matskevich/item-based-recommendations/internal/api/utils"
	"github.com/st-matskevich/item-based-recommendations/internal/firebase"
)

type BaseHandler func(*http.Request) utils.HandlerResponse

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc BaseHandler
}

func addCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
}

func HandleCORS(r *http.Request) utils.HandlerResponse {
	return utils.MakeHandlerResponse(http.StatusOK, struct{}{}, nil)
}

func AuthMiddleware(inner BaseHandler) BaseHandler {
	return BaseHandler(func(r *http.Request) utils.HandlerResponse {
		uid, err := firebase.GetFirebaseAuth().Verify(r.Header.Get("Authorization"))
		if err != nil {
			return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.AUTHORIZATION_ERROR), err)
		}
		ctx := context.WithValue(r.Context(), utils.USER_ID_CTX_KEY, uid)

		return inner(r.WithContext(ctx))
	})
}

func BaseMiddleware(inner BaseHandler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		//TODO: this should be removed in prod
		addCORSHeaders(w)

		response := inner(r)

		log.Printf(
			"%-8s %-64s %d %-32s %-10s",
			r.Method,
			r.RequestURI,
			response.Code,
			name,
			time.Since(start),
		)
		if response.Err != nil {
			log.Printf("%s error: %v", name, response.Err)
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(response.Code)
		err := json.NewEncoder(w).Encode(response.Response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("%s response encoding error: %v", name, err)
		}
	})
}

func MakeRouter() *mux.Router {

	router := mux.NewRouter().StrictSlash(true)

	//TODO: this should be removed in prod
	router.Methods("OPTIONS").Handler(BaseMiddleware(HandleCORS, "CORS"))

	for _, route := range routes {
		handler := BaseMiddleware(route.HandlerFunc, route.Name)
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}

var routes = []Route{
	{
		"Get Recommendations",
		"GET",
		"/recommendations",
		AuthMiddleware(similarity.HandleGetRecommendations),
	},
	{
		"Get User Profile",
		"GET",
		"/profile",
		AuthMiddleware(profile.HandleGetUserProfile),
	},
	{
		"Set User Profile",
		"POST",
		"/profile",
		AuthMiddleware(profile.HandleSetUserProfile),
	},
	{
		"Get Tasks Feed",
		"GET",
		"/tasks",
		AuthMiddleware(tasks.HandleGetTasksFeed),
	},
	{
		"Get Task",
		"GET",
		"/tasks/{task}",
		AuthMiddleware(tasks.HandleGetTask),
	},
	{
		"Create Task",
		"POST",
		"/tasks",
		AuthMiddleware(tasks.HandleCreateTask),
	},
	{
		"Get Replies",
		"GET",
		"/tasks/{task}/replies",
		AuthMiddleware(replies.HandleGetReplies),
	},
	{
		"Create Reply",
		"POST",
		"/tasks/{task}/replies",
		AuthMiddleware(replies.HandleCreateReply),
	},
	{
		"Set Task Doer",
		"POST",
		"/tasks/{task}/doer",
		AuthMiddleware(tasks.HandleSetDoer),
	},
}
