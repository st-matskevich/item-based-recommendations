package profile

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/st-matskevich/item-based-recommendations/internal/api/utils"
	"github.com/st-matskevich/item-based-recommendations/internal/db"
	"github.com/st-matskevich/item-based-recommendations/internal/firebase"
)

type UserProfile struct {
	Name       string `json:"name"`
	IsCustomer bool   `json:"customer"`
}

func getUserProfileReader(client *db.SQLClient, userID int64) (db.ResponseReader, error) {
	return client.Query("SELECT name, is_customer FROM users WHERE user_id = $1", userID)
}

func getUserProfile(reader db.ResponseReader) (UserProfile, error) {
	result := UserProfile{}
	found, err := reader.Next(&result.Name, &result.IsCustomer)
	if !found && err == nil {
		err = errors.New(utils.SQL_NO_RESULT)
	}
	return result, err
}

func HandleGetUserProfile(w http.ResponseWriter, r *http.Request) utils.HandlerResponse {
	uid, err := firebase.GetFirebaseAuth().Verify(r.Header.Get("Authorization"))
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.AUTHORIZATION_ERROR), err)
	}

	reader, err := getUserProfileReader(db.GetSQLClient(), uid)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}
	defer reader.Close()

	profile, err := getUserProfile(reader)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	return utils.MakeHandlerResponse(http.StatusOK, profile, nil)
}

func setUserProfile(client *db.SQLClient, profile UserProfile, userID int64) error {
	reader, err := client.Query("UPDATE users SET name = $2, is_customer = $3 WHERE user_id = $1", userID, profile.Name, profile.IsCustomer)
	reader.Close()
	return err
}

func parseUserProfile(profile UserProfile) error {
	if profile.Name == "" {
		return errors.New(utils.INVALID_INPUT)
	}

	if len([]rune(profile.Name)) > 32 {
		return errors.New(utils.INVALID_INPUT)
	}

	return nil
}

func HandleSetUserProfile(w http.ResponseWriter, r *http.Request) utils.HandlerResponse {
	uid, err := firebase.GetFirebaseAuth().Verify(r.Header.Get("Authorization"))
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.AUTHORIZATION_ERROR), err)
	}

	input := UserProfile{}
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.DECODER_ERROR), err)
	}

	err = parseUserProfile(input)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.DECODER_ERROR), err)
	}

	err = setUserProfile(db.GetSQLClient(), input, uid)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	return utils.MakeHandlerResponse(http.StatusOK, struct{}{}, nil)
}
