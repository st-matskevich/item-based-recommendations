package api

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type HandlerResponse struct {
	code int
	err  error
}
type Handler func(http.ResponseWriter, *http.Request) HandlerResponse

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc Handler
}

func Logger(inner Handler, name string) http.Handler {
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
	})
}

func MakeRouter() *mux.Router {

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {

		handler := Logger(route.HandlerFunc, route.Name)

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
		"/recommendations/{id}",
		similarityRequest,
	},
}
