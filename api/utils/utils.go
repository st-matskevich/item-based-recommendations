package utils

import "net/http"

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

func CORSHandler(w http.ResponseWriter, r *http.Request) HandlerResponse {
	return HandlerResponse{http.StatusOK, struct{}{}, nil}
}
