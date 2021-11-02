package controller

import (
	"net/http"

	"github.com/st-matskevich/item-based-recommendations/internal/api/middleware"
	"github.com/st-matskevich/item-based-recommendations/internal/api/repository"
	"github.com/st-matskevich/item-based-recommendations/internal/api/utils"
)

type NotificationsController struct {
	NotificationsRepo repository.NotificationsRepository
	RepliesRepo       repository.RepliesRepository
	TasksRepo         repository.TasksRepository
}

func (controller *NotificationsController) GetRoutes() []utils.Route {
	return []utils.Route{
		{
			Name:    "Get Notifications",
			Method:  "GET",
			Pattern: "/notifications",
			Handler: middleware.AuthMiddleware(controller.HandleGetNotifications),
		},
	}
}

func (controller *NotificationsController) HandleGetNotifications(r *http.Request) utils.HandlerResponse {
	uid := utils.GetUserID(r.Context())

	notifications, err := controller.NotificationsRepo.GetNotifications(uid)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	for idx, val := range notifications {
		switch val.Type {
		case repository.NEW_REPLY_NOTIFICATION:
			reply, err := controller.RepliesRepo.GetReply(val.TriggerID)
			if err != nil {
				return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
			}
			notifications[idx].Content = reply
		case repository.TASK_CLOSE_NOTIFICATION:
			task, err := controller.TasksRepo.GetTask(val.TriggerID)
			if err != nil {
				return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
			}
			notifications[idx].Content = task
		}
	}

	return utils.MakeHandlerResponse(http.StatusOK, notifications, nil)
}
