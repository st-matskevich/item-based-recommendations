package controller

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/st-matskevich/item-based-recommendations/internal/api/middleware"
	"github.com/st-matskevich/item-based-recommendations/internal/api/repository"
	"github.com/st-matskevich/item-based-recommendations/internal/api/utils"
)

type TaskReplies struct {
	User *repository.Reply  `json:"user"`
	Doer *repository.Reply  `json:"doer"`
	All  []repository.Reply `json:"all"`
}

type RepliesController struct {
	RepliesRepo       repository.RepliesRepository
	TasksRepo         repository.TasksRepository
	NotificationsRepo repository.NotificationsRepository
}

func (controller *RepliesController) GetRoutes() []utils.Route {
	return []utils.Route{
		{
			Name:    "Get Replies",
			Method:  "GET",
			Pattern: "/tasks/{task}/replies",
			Handler: middleware.AuthMiddleware(controller.HandleGetReplies),
		},
		{
			Name:    "Create Reply",
			Method:  "POST",
			Pattern: "/tasks/{task}/replies",
			Handler: middleware.AuthMiddleware(controller.HandleCreateReply),
		},
		{
			Name:    "Hide Reply",
			Method:  "DELETE",
			Pattern: "/tasks/{task}/replies/{reply}",
			Handler: middleware.AuthMiddleware(controller.HandleHideReply),
		},
	}
}

func validateReply(reply repository.Reply) error {
	if reply.Text == "" {
		return errors.New(utils.INVALID_INPUT)
	}

	if len([]rune(reply.Text)) > 512 {
		return errors.New(utils.INVALID_INPUT)
	}

	return nil
}

func (controller *RepliesController) HandleGetReplies(r *http.Request) utils.HandlerResponse {
	uid := utils.GetUserID(r.Context())
	result := TaskReplies{}

	taskID, err := utils.UIDFromString(mux.Vars(r)["task"])
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.DECODER_ERROR), err)
	}

	task, err := controller.TasksRepo.GetTask(taskID)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	result.User, err = controller.RepliesRepo.GetUserReply(taskID, uid)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	if task.Customer.ID == uid {
		result.Doer, err = controller.RepliesRepo.GetDoerReply(taskID)
		if err != nil {
			return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
		}
	}

	if task.Customer.ID == uid {
		result.All, err = controller.RepliesRepo.GetReplies(taskID)
		if err != nil {
			return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
		}
	}

	return utils.MakeHandlerResponse(http.StatusOK, result, nil)
}

func (controller *RepliesController) HandleCreateReply(r *http.Request) utils.HandlerResponse {
	uid := utils.GetUserID(r.Context())

	input := repository.Reply{}
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.DECODER_ERROR), err)
	}
	input.Creator.ID = uid

	err = validateReply(input)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.BAD_INPUT), err)
	}

	taskID, err := utils.UIDFromString(mux.Vars(r)["task"])
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.DECODER_ERROR), err)
	}

	replyID, err := controller.RepliesRepo.CreateReply(taskID, input)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	task, err := controller.TasksRepo.GetTask(taskID)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	err = controller.NotificationsRepo.CreateNotification(task.Customer.ID, repository.NEW_REPLY_NOTIFICATION, replyID)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	return utils.MakeHandlerResponse(http.StatusOK, struct{}{}, nil)
}

func (controller *RepliesController) HandleHideReply(r *http.Request) utils.HandlerResponse {
	uid := utils.GetUserID(r.Context())

	taskID, err := utils.UIDFromString(mux.Vars(r)["task"])
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.DECODER_ERROR), err)
	}

	replyID, err := utils.UIDFromString(mux.Vars(r)["reply"])
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.DECODER_ERROR), err)
	}

	task, err := controller.TasksRepo.GetTask(taskID)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	if task.Customer.ID != uid {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.AUTHORIZATION_ERROR), errors.New(utils.INSUFFICIENT_RIGHTS))
	}

	err = controller.RepliesRepo.HideReply(replyID)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	return utils.MakeHandlerResponse(http.StatusOK, struct{}{}, nil)
}
