package utils

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
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

func HandleCORS(w http.ResponseWriter, r *http.Request) HandlerResponse {
	return HandlerResponse{http.StatusOK, struct{}{}, nil}
}

type UID int64

func (val UID) MarshalJSON() ([]byte, error) {
	str := strconv.FormatInt(int64(val), 10)
	enc := base64.StdEncoding.EncodeToString([]byte(str))
	json, err := json.Marshal(enc)
	return json, err
}

func (val *UID) UnmarshalJSON(data []byte) error {
	var enc string
	if err := json.Unmarshal(data, &enc); err != nil {
		return err
	}

	return val.FromString(enc)
}

func (val *UID) FromString(enc string) error {
	str, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		return err
	}

	i, err := strconv.ParseInt(string(str), 10, 64)
	if err != nil {
		return err
	}

	*val = UID(i)
	return nil
}

type UserData struct {
	ID   UID    `json:"id"`
	Name string `json:"name"`
}
