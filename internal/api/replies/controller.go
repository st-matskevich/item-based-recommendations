package replies

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/st-matskevich/item-based-recommendations/internal/api/middleware"
	"github.com/st-matskevich/item-based-recommendations/internal/api/tasks"
	"github.com/st-matskevich/item-based-recommendations/internal/api/utils"
)

type TaskReplies struct {
	User *Reply  `json:"user"`
	Doer *Reply  `json:"doer"`
	All  []Reply `json:"all"`
}

type RepliesController struct {
	RepliesRepo RepliesRepository
	TasksRepo   tasks.TasksRepository
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
	}
}

func validateReply(reply Reply) error {
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

	input := Reply{}
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.DECODER_ERROR), err)
	}
	input.Creator.ID = uid

	err = validateReply(input)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.DECODER_ERROR), err)
	}

	taskID, err := utils.UIDFromString(mux.Vars(r)["task"])
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.DECODER_ERROR), err)
	}

	err = controller.RepliesRepo.CreateReply(taskID, input)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	return utils.MakeHandlerResponse(http.StatusOK, struct{}{}, nil)
}
