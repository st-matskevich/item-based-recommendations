package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type HandlerResponse struct {
	code     int
	response interface{}
	err      error
}
type Handler func(http.ResponseWriter, *http.Request) HandlerResponse

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc Handler
}

type ErrorResponse struct {
	Error ErrorMessage `json:"error"`
}
type ErrorMessage struct {
	Code string `json:"code"`
}

func CreateErrorMessage(code string) ErrorResponse {
	return ErrorResponse{ErrorMessage{code}}
}

func BaseHandler(inner Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		response := inner(w, r)

		log.Printf(
			"%s\t%s\t%d\t%s\t%s",
			r.Method,
			r.RequestURI,
			response.code,
			name,
			time.Since(start),
		)
		if response.err != nil {
			log.Printf("%s error: %v", name, response.err)
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(response.code)
		err := json.NewEncoder(w).Encode(response.response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("%s response encoding error: %v", name, err)
		}
	})
}

func MakeRouter() *mux.Router {

	router := mux.NewRouter().StrictSlash(true)
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
		"Recommendations",
		"GET",
		"/recommendations",
		similarityRequest,
	},
}
