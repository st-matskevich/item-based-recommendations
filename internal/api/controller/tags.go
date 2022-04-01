package controller

import (
	"errors"
	"net/http"

	"github.com/st-matskevich/item-based-recommendations/internal/api/middleware"
	"github.com/st-matskevich/item-based-recommendations/internal/api/repository"
	"github.com/st-matskevich/item-based-recommendations/internal/api/utils"
)

type TagsController struct {
	TagsRepo repository.TagsRepository
}

type TagsSearchRequest struct {
	Request string `json:"request"`
}

func (controller *TagsController) GetRoutes() []utils.Route {
	return []utils.Route{
		{
			Name:    "Search tags",
			Method:  "GET",
			Pattern: "/tags",
			Handler: middleware.AuthMiddleware(controller.HandleSearchTags),
		},
	}
}

func (controller *TagsController) HandleSearchTags(r *http.Request) utils.HandlerResponse {
	request := r.FormValue("query")

	if request == "" {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.BAD_INPUT), errors.New(utils.INVALID_INPUT))
	}

	if len([]rune(request)) > 32 {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.BAD_INPUT), errors.New(utils.INVALID_INPUT))
	}

	tags, err := controller.TagsRepo.SearchTags(request)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.DECODER_ERROR), err)
	}

	return utils.MakeHandlerResponse(http.StatusOK, tags, nil)
}
