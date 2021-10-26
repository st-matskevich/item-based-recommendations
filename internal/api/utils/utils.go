package utils

import (
	"context"
	"encoding/base64"
	"encoding/json"
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
	ID         UID    `json:"id"`
	Name       string `json:"name"`
	IsCustomer *bool  `json:"customer,omitempty"`
}

type ContextKey int

const (
	USER_ID_CTX_KEY ContextKey = 0
)

func GetUserID(ctx context.Context) UID {
	return ctx.Value(USER_ID_CTX_KEY).(UID)
}
