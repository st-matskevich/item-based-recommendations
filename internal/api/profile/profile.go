package profile

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/st-matskevich/item-based-recommendations/internal/api/utils"
	"github.com/st-matskevich/item-based-recommendations/internal/db"
)

func getUserProfileReader(client *db.SQLClient, userID utils.UID) (db.ResponseReader, error) {
	return client.Query("SELECT name, is_customer FROM users WHERE user_id = $1", userID)
}

func getUserProfile(reader db.ResponseReader) (utils.UserData, error) {
	result := utils.UserData{}
	err := reader.GetRow(&result.Name, &result.IsCustomer)
	return result, err
}

func HandleGetUserProfile(r *http.Request) utils.HandlerResponse {
	uid := utils.GetUserID(r.Context())

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

func setUserProfile(client *db.SQLClient, profile utils.UserData, userID utils.UID) error {
	return client.Exec("UPDATE users SET name = $2, is_customer = $3 WHERE user_id = $1", userID, profile.Name, profile.IsCustomer)
}

func parseUserProfile(profile utils.UserData) error {
	if profile.Name == "" {
		return errors.New(utils.INVALID_INPUT)
	}

	if len([]rune(profile.Name)) > 32 {
		return errors.New(utils.INVALID_INPUT)
	}

	return nil
}

func HandleSetUserProfile(r *http.Request) utils.HandlerResponse {
	uid := utils.GetUserID(r.Context())

	input := utils.UserData{}
	err := json.NewDecoder(r.Body).Decode(&input)
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
