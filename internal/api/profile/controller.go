package profile

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/st-matskevich/item-based-recommendations/internal/api/middleware"
	"github.com/st-matskevich/item-based-recommendations/internal/api/utils"
)

type ProfileController struct {
	ProfileRepo ProfileRepository
}

func (controller *ProfileController) GetRoutes() []utils.Route {
	return []utils.Route{
		{
			Name:    "Get User Profile",
			Method:  "GET",
			Pattern: "/profile",
			Handler: middleware.AuthMiddleware(controller.HandleGetUserProfile),
		},
		{
			Name:    "Set User Profile",
			Method:  "POST",
			Pattern: "/profile",
			Handler: middleware.AuthMiddleware(controller.HandleSetUserProfile),
		},
	}
}

func validateProfile(profile UserData) error {
	if profile.Name == "" {
		return errors.New(utils.INVALID_INPUT)
	}

	if len([]rune(profile.Name)) > 32 {
		return errors.New(utils.INVALID_INPUT)
	}

	return nil
}

func (controller *ProfileController) HandleGetUserProfile(r *http.Request) utils.HandlerResponse {
	uid := utils.GetUserID(r.Context())

	profile, err := controller.ProfileRepo.GetProfile(uid)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	return utils.MakeHandlerResponse(http.StatusOK, profile, nil)
}

func (controller *ProfileController) HandleSetUserProfile(r *http.Request) utils.HandlerResponse {
	uid := utils.GetUserID(r.Context())

	input := UserData{}
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.DECODER_ERROR), err)
	}

	err = validateProfile(input)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.DECODER_ERROR), err)
	}

	err = controller.ProfileRepo.SetProfile(uid, input)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	return utils.MakeHandlerResponse(http.StatusOK, struct{}{}, nil)
}
