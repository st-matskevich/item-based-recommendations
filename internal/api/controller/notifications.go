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

type NewReplyNotificationContent struct {
	Task  repository.Task  `json:"task"`
	Reply repository.Reply `json:"reply"`
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

	return utils.MakeHandlerResponse(http.StatusOK, notifications, nil)
}
