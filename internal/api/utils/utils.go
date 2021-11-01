package utils

import (
	"net/http"
)

type ErrorMessage struct {
	Code string `json:"code"`
}

func MakeErrorMessage(code string) ErrorMessage {
	return ErrorMessage{code}
}

type HandlerResponse struct {
	Code     int
	Response interface{}
	Err      error
}

func MakeHandlerResponse(code int, response interface{}, err error) HandlerResponse {
	return HandlerResponse{code, response, err}
}

type BaseHandler func(*http.Request) HandlerResponse

type Route struct {
	Name    string
	Method  string
	Pattern string
	Handler BaseHandler
}

type Controller interface {
	GetRoutes() []Route
}
