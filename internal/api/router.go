package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/st-matskevich/item-based-recommendations/internal/api/profile"
	"github.com/st-matskevich/item-based-recommendations/internal/api/replies"
	"github.com/st-matskevich/item-based-recommendations/internal/api/tasks"
	"github.com/st-matskevich/item-based-recommendations/internal/api/utils"
	"github.com/st-matskevich/item-based-recommendations/internal/db"
)

func addCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
}

func HandleCORS(r *http.Request) utils.HandlerResponse {
	return utils.MakeHandlerResponse(http.StatusOK, struct{}{}, nil)
}

func BaseMiddleware(inner utils.BaseHandler, name string) http.Handler {
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

	controllers := []utils.Controller{
		&profile.ProfileController{
			ProfileRepo: &profile.ProfileSQLRepository{
				SQLClient: db.GetSQLClient(),
			},
		},
		&tasks.TasksController{
			TasksRepo: &tasks.TasksSQLRepository{
				SQLClient: db.GetSQLClient(),
			},
		},
		&replies.RepliesController{
			RepliesRepo: &replies.RepliesSQLRepository{
				SQLClient: db.GetSQLClient(),
			},
			TasksRepo: &tasks.TasksSQLRepository{
				SQLClient: db.GetSQLClient(),
			},
		},
	}

	//TODO: this should be removed in prod
	router.Methods("OPTIONS").Handler(BaseMiddleware(HandleCORS, "CORS"))

	for _, controller := range controllers {
		for _, route := range controller.GetRoutes() {
			handler := BaseMiddleware(route.Handler, route.Name)
			router.
				Methods(route.Method).
				Path(route.Pattern).
				Name(route.Name).
				Handler(handler)
		}
	}

	return router
}
