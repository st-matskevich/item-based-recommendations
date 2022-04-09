package controller

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/st-matskevich/item-based-recommendations/internal/api/middleware"
	"github.com/st-matskevich/item-based-recommendations/internal/api/repository"
	"github.com/st-matskevich/item-based-recommendations/internal/api/utils"
)

type TasksController struct {
	TasksRepo         repository.TasksRepository
	ProfileRepo       repository.ProfileRepository
	TagsRepo          repository.TagsRepository
	RepliesRepo       repository.RepliesRepository
	NotificationsRepo repository.NotificationsRepository
}

type TaskWrapper struct {
	*repository.Task
	Tags         []repository.Tag `json:"tags"`
	Owns         bool             `json:"owns"`
	Liked        bool             `json:"liked"`
	RepliesCount int32            `json:"repliesCount"`
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
			Name:    "Like Task",
			Method:  "POST",
			Pattern: "/tasks/{task}/like",
			Handler: middleware.AuthMiddleware(controller.LikeTask),
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

func (controller *TasksController) buildTaskWrapper(uid utils.UID, task *repository.Task) (*TaskWrapper, error) {
	var err error
	wrapper := TaskWrapper{}
	wrapper.Task = task
	wrapper.Owns = uid == wrapper.Customer.ID

	wrapper.RepliesCount, err = controller.RepliesRepo.GetRepliesCount(wrapper.ID)
	if err != nil {
		return nil, err
	}

	wrapper.Tags, err = controller.TagsRepo.GetTaskTags(wrapper.ID)
	if err != nil {
		return nil, err
	}

	wrapper.Liked, err = controller.TasksRepo.IsTaskLiked(uid, wrapper.ID)
	if err != nil {
		return nil, err
	}

	return &wrapper, nil
}

func (controller *TasksController) HandleGetTasksFeed(r *http.Request) utils.HandlerResponse {
	var err error
	var tasks []repository.Task

	uid := utils.GetUserID(r.Context())

	scope := r.FormValue("scope")
	query := r.FormValue("query")

	if scope == repository.RECOMMENDATIONS {
		tasks, err = controller.GetRecommendations(uid)
	} else {
		tasks, err = controller.TasksRepo.GetTasksFeed(scope, query, uid)
	}

	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	result := []TaskWrapper{}
	for idx := range tasks {
		wrapper, err := controller.buildTaskWrapper(uid, &tasks[idx])
		if err != nil {
			return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
		}
		result = append(result, *wrapper)
	}

	return utils.MakeHandlerResponse(http.StatusOK, result, nil)
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

	wrapper, err := controller.buildTaskWrapper(uid, task)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	return utils.MakeHandlerResponse(http.StatusOK, wrapper, nil)
}

func (controller *TasksController) LikeTask(r *http.Request) utils.HandlerResponse {
	uid := utils.GetUserID(r.Context())

	taskID, err := utils.UIDFromString(mux.Vars(r)["task"])
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.DECODER_ERROR), err)
	}

	likes, err := strconv.ParseBool(r.FormValue("value"))
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.DECODER_ERROR), err)
	}

	err = controller.TasksRepo.SetTaskLike(uid, taskID, likes)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	return utils.MakeHandlerResponse(http.StatusOK, likes, nil)
}

func validateTask(task *TaskWrapper) error {
	if task.Name == "" || task.Description == "" {
		return errors.New(utils.INVALID_INPUT)
	}

	if len([]rune(task.Name)) > 128 {
		return errors.New(utils.INVALID_INPUT)
	}

	if len([]rune(task.Description)) > 2048 {
		return errors.New(utils.INVALID_INPUT)
	}

	if len(task.Tags) > 5 {
		return errors.New(utils.INVALID_INPUT)
	}

	if len(task.Tags) < 1 {
		return errors.New(utils.INVALID_INPUT)
	}

	for _, tag := range task.Tags {
		if len([]rune(tag.Text)) > 32 {
			return errors.New(utils.INVALID_INPUT)
		}
	}

	return nil
}

func (controller *TasksController) HandleCreateTask(r *http.Request) utils.HandlerResponse {
	uid := utils.GetUserID(r.Context())

	input := TaskWrapper{}
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.DECODER_ERROR), err)
	}
	input.Customer.ID = uid

	err = validateTask(&input)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.BAD_INPUT), err)
	}

	taskID, err := controller.TasksRepo.CreateTask(*input.Task)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	for _, tag := range input.Tags {
		if tag.ID == 0 {
			tag.Text = strings.ToLower(tag.Text)
			tag.ID, err = controller.TagsRepo.CreateTag(tag.Text)
			if err != nil {
				return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
			}
		}

		err = controller.TagsRepo.AddTagToTask(taskID, tag.ID)
		if err != nil {
			return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
		}
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

func normalizeVector(vector map[utils.UID]float32) {
	magnitude := float32(0)
	for _, val := range vector {
		magnitude += val * val
	}
	magnitude = float32(math.Sqrt(float64(magnitude)))

	for id, val := range vector {
		vector[id] = val / magnitude
	}
}

func buildUserRecommendationsVector(userTags []repository.TaskTagLink) map[utils.UID]float32 {
	result := map[utils.UID]float32{}
	uniqueTasks := map[utils.UID]struct{}{}

	for _, row := range userTags {
		result[row.TagID] += 1
		uniqueTasks[row.TaskID] = struct{}{}
	}

	for tagID := range result {
		result[tagID] /= float32(len(uniqueTasks))
	}

	normalizeVector(result)

	return result
}

func buildTasksRecommendationsVector(tasksTags []repository.TaskTagLink) map[utils.UID]map[utils.UID]float32 {
	result := map[utils.UID]map[utils.UID]float32{}
	uniqueTasks := map[utils.UID]struct{}{}
	uniqueTags := map[utils.UID]float32{}

	for _, row := range tasksTags {
		if _, contains := result[row.TaskID]; !contains {
			result[row.TaskID] = map[utils.UID]float32{}
		}

		result[row.TaskID][row.TagID] = 0
		uniqueTags[row.TagID] += 1
		uniqueTasks[row.TaskID] = struct{}{}
	}

	for taskID, tagsMap := range result {
		for tagID := range tagsMap {
			result[taskID][tagID] = uniqueTags[tagID] / float32(len(uniqueTasks))
		}

		normalizeVector(result[taskID])
	}

	return result
}

func getRecommendedTasks(userVector map[utils.UID]float32, tasksVectors map[utils.UID]map[utils.UID]float32, threshold float32) []utils.UID {
	result := []utils.UID{}
	similarity := float32(0)

	for taskID, tagsMap := range tasksVectors {
		//TODO: use go coroutines
		similarity = 0
		for tagID, tagWeight := range tagsMap {
			if val, ok := userVector[tagID]; ok {
				similarity += tagWeight * val
			}
		}

		if similarity >= threshold {
			result = append(result, taskID)
		}
	}
	return result
}

func (controller *TasksController) GetRecommendations(userID utils.UID) ([]repository.Task, error) {
	threshold, err := strconv.ParseFloat(os.Getenv("SIMILARITY_THRESHOLD"), 32)
	if err != nil {
		return nil, err
	}

	userTags, err := controller.ProfileRepo.GetLikedTags(userID)
	if err != nil {
		return nil, err
	}

	tasksTags, err := controller.TasksRepo.GetTasksTags(userID)
	if err != nil {
		return nil, err
	}

	userVector := buildUserRecommendationsVector(userTags)
	tasksVector := buildTasksRecommendationsVector(tasksTags)

	recommendedTasks := getRecommendedTasks(userVector, tasksVector, float32(threshold))
	result := []repository.Task{}
	for _, taskID := range recommendedTasks {
		task, err := controller.TasksRepo.GetTask(taskID)
		if err != nil {
			return nil, err
		}

		result = append(result, *task)
	}

	return result, nil
}
