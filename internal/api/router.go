package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/st-matskevich/item-based-recommendations/internal/api/profile"
	"github.com/st-matskevich/item-based-recommendations/internal/api/similarity"
	"github.com/st-matskevich/item-based-recommendations/internal/api/tasks"
	"github.com/st-matskevich/item-based-recommendations/internal/api/utils"
)

type Handler func(http.ResponseWriter, *http.Request) utils.HandlerResponse

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc Handler
}

func addCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
}

func BaseHandler(inner Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		//TODO: this should be removed in prod
		addCORSHeaders(w)

		response := inner(w, r)

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
	router.Methods("OPTIONS").Handler(BaseHandler(utils.HandleCORS, "CORS"))

	for _, route := range routes {

		handler := BaseHandler(route.HandlerFunc, route.Name)

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
		similarity.HandleGetRecommendations,
	},
	{
		"Get User Profile",
		"GET",
		"/profile",
		profile.HandleGetUserProfile,
	},
	{
		"Set User Profile",
		"POST",
		"/profile",
		profile.HandleSetUserProfile,
	},
	{
		"Get Tasks",
		"GET",
		"/tasks",
		tasks.HandleGetTasks,
	},
	{
		"Create Task",
		"POST",
		"/tasks",
		tasks.HandleCreateTask,
	},
}
