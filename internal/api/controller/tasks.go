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

type TasksController struct {
	TasksRepo         repository.TasksRepository
	NotificationsRepo repository.NotificationsRepository
}

func (controller *TasksController) GetRoutes() []utils.Route {
	return []utils.Route{
		{
			Name:    "Get Tasks Feed",
			Method:  "GET",
			Pattern: "/tasks",
			Handler: middleware.AuthMiddleware(controller.HandleGetTasksFeed),
		},
		{
			Name:    "Get Task",
			Method:  "GET",
			Pattern: "/tasks/{task}",
			Handler: middleware.AuthMiddleware(controller.HandleGetTask),
		},
		{
			Name:    "Create Task",
			Method:  "POST",
			Pattern: "/tasks",
			Handler: middleware.AuthMiddleware(controller.HandleCreateTask),
		},
		{
			Name:    "Close Task",
			Method:  "POST",
			Pattern: "/tasks/{task}/close",
			Handler: middleware.AuthMiddleware(controller.HandleCloseTask),
		},
	}
}

func (controller *TasksController) HandleGetTasksFeed(r *http.Request) utils.HandlerResponse {
	uid := utils.GetUserID(r.Context())

	scope := r.FormValue("scope")

	tasks, err := controller.TasksRepo.GetTasksFeed(scope, uid)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	for idx, val := range tasks {
		tasks[idx].Owns = uid == val.Customer.ID
	}

	return utils.MakeHandlerResponse(http.StatusOK, tasks, nil)
}

func (controller *TasksController) HandleGetTask(r *http.Request) utils.HandlerResponse {
	uid := utils.GetUserID(r.Context())

	taskID, err := utils.UIDFromString(mux.Vars(r)["task"])
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.DECODER_ERROR), err)
	}

	task, err := controller.TasksRepo.GetTask(taskID)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	task.Owns = uid == task.Customer.ID

	return utils.MakeHandlerResponse(http.StatusOK, task, nil)
}

func validateTask(task repository.Task) error {
	if task.Name == "" || task.Description == "" {
		return errors.New(utils.INVALID_INPUT)
	}

	if len([]rune(task.Name)) > 128 {
		return errors.New(utils.INVALID_INPUT)
	}

	if len([]rune(task.Description)) > 2048 {
		return errors.New(utils.INVALID_INPUT)
	}

	return nil
}

func (controller *TasksController) HandleCreateTask(r *http.Request) utils.HandlerResponse {
	uid := utils.GetUserID(r.Context())

	input := repository.Task{}
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.DECODER_ERROR), err)
	}
	input.Customer.ID = uid

	err = validateTask(input)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.DECODER_ERROR), err)
	}

	err = controller.TasksRepo.CreateTask(input)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	return utils.MakeHandlerResponse(http.StatusOK, struct{}{}, nil)
}

func (controller *TasksController) HandleCloseTask(r *http.Request) utils.HandlerResponse {
	uid := utils.GetUserID(r.Context())

	doer := repository.UserData{}
	err := json.NewDecoder(r.Body).Decode(&doer)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.DECODER_ERROR), err)
	}

	taskID, err := utils.UIDFromString(mux.Vars(r)["task"])
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

	err = controller.TasksRepo.CloseTask(taskID, doer.ID)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	err = controller.NotificationsRepo.CreateNotification(doer.ID, repository.TASK_CLOSE_NOTIFICATION, taskID)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	return utils.MakeHandlerResponse(http.StatusOK, struct{}{}, nil)
}
