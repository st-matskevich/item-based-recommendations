package tasks

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/st-matskevich/item-based-recommendations/internal/api/utils"
	"github.com/st-matskevich/item-based-recommendations/internal/db"
	"github.com/st-matskevich/item-based-recommendations/internal/firebase"
)

const (
	NOT_ASSIGNED = "NOT_ASSIGNED"
	CUSTOMER     = "CUSTOMER"
	DOER         = "DOER"
)

type Task struct {
	ID           utils.UID      `json:"id"`
	Name         string         `json:"name"`
	Description  string         `json:"description,omitempty"`
	Customer     utils.UserData `json:"customer"`
	Owns         bool           `json:"owns"`
	Closed       bool           `json:"closed"`
	RepliesCount int32          `json:"repliesCount"`
	CreatedAt    time.Time      `json:"createdAt"`
}

func getTasksFeedReader(client *db.SQLClient, userID utils.UID, scope string) (db.ResponseReader, error) {
	switch scope {
	case CUSTOMER:
		return client.Query(`SELECT tasks.task_id, tasks.name, tasks.doer_id IS NOT NULL, users.user_id, users.name, COUNT(replies.task_id), tasks.created_at
							FROM tasks 
							JOIN users 
							ON tasks.customer_id = users.user_id
							AND tasks.customer_id = $1
							LEFT JOIN replies
							ON tasks.task_id = replies.task_id
							GROUP BY tasks.task_id, tasks.name, tasks.doer_id, users.user_id, users.name, tasks.created_at
							ORDER BY tasks.task_id`, userID)
	case DOER:
		return client.Query(`SELECT tasks.task_id, tasks.name, tasks.doer_id IS NOT NULL, users.user_id, users.name, COUNT(replies.task_id), tasks.created_at
							FROM tasks 
							JOIN users 
							ON tasks.customer_id = users.user_id
							AND tasks.doer_id = $1
							LEFT JOIN replies
							ON tasks.task_id = replies.task_id
							GROUP BY tasks.task_id, tasks.name, tasks.doer_id, users.user_id, users.name, tasks.created_at
							ORDER BY tasks.task_id`, userID)
	}
	return client.Query(`SELECT tasks.task_id, tasks.name, tasks.doer_id IS NOT NULL, users.user_id, users.name, COUNT(replies.task_id), tasks.created_at
						FROM tasks 
						JOIN users 
						ON tasks.customer_id = users.user_id
						AND tasks.doer_id IS NULL
						LEFT JOIN replies
						ON tasks.task_id = replies.task_id
						GROUP BY tasks.task_id, tasks.name, tasks.doer_id, users.user_id, users.name, tasks.created_at
						ORDER BY tasks.task_id`)
}

func getTasksFeed(reader db.ResponseReader, userID utils.UID) ([]Task, error) {
	result := []Task{}
	row := Task{}
	ok, err := reader.NextRow(&row.ID, &row.Name, &row.Closed, &row.Customer.ID, &row.Customer.Name, &row.RepliesCount, &row.CreatedAt)
	for ; ok; ok, err = reader.NextRow(&row.ID, &row.Name, &row.Closed, &row.Customer.ID, &row.Customer.Name, &row.RepliesCount, &row.CreatedAt) {
		row.Owns = userID == row.Customer.ID
		result = append(result, row)
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func HandleGetTasksFeed(w http.ResponseWriter, r *http.Request) utils.HandlerResponse {
	uid, err := firebase.GetFirebaseAuth().Verify(r.Header.Get("Authorization"))
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.AUTHORIZATION_ERROR), err)
	}

	scope := r.FormValue("scope")
	reader, err := getTasksFeedReader(db.GetSQLClient(), uid, scope)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}
	defer reader.Close()

	tasks, err := getTasksFeed(reader, uid)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	return utils.MakeHandlerResponse(http.StatusOK, tasks, nil)
}

func getTaskReader(client *db.SQLClient, taskID utils.UID) (db.ResponseReader, error) {
	return client.Query(`SELECT tasks.task_id, tasks.name, tasks.description, tasks.doer_id IS NOT NULL, users.user_id, users.name, COUNT(replies.task_id), tasks.created_at  
						FROM tasks 
						JOIN users 
						ON tasks.customer_id = users.user_id
						AND tasks.task_id = $1
						LEFT JOIN replies
						ON tasks.task_id = replies.task_id
						GROUP BY tasks.task_id, tasks.name, tasks.description, tasks.doer_id, users.user_id, users.name, tasks.created_at
						ORDER BY tasks.task_id`, taskID)
}

func getTask(reader db.ResponseReader, userID utils.UID) (Task, error) {
	result := Task{}
	err := reader.GetRow(&result.ID, &result.Name, &result.Description, &result.Closed, &result.Customer.ID, &result.Customer.Name, &result.RepliesCount, &result.CreatedAt)

	if err != nil {
		return result, err
	}

	result.Owns = userID == result.Customer.ID
	return result, err
}

func HandleGetTask(w http.ResponseWriter, r *http.Request) utils.HandlerResponse {
	uid, err := firebase.GetFirebaseAuth().Verify(r.Header.Get("Authorization"))
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.AUTHORIZATION_ERROR), err)
	}

	var taskID utils.UID
	err = taskID.FromString(mux.Vars(r)["task"])
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.DECODER_ERROR), err)
	}

	reader, err := getTaskReader(db.GetSQLClient(), taskID)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}
	defer reader.Close()

	task, err := getTask(reader, uid)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	return utils.MakeHandlerResponse(http.StatusOK, task, nil)
}

func createTask(client *db.SQLClient, task Task, userID utils.UID) error {
	reader, err := client.Query("INSERT INTO tasks(name, description, customer_id) VALUES ($1, $2, $3)", task.Name, task.Description, userID)
	reader.Close()
	return err
}

func parseTask(task Task) error {
	if task.Name == "" || task.Description == "" {
		return errors.New(utils.INVALID_INPUT)
	}

	if len([]rune(task.Name)) > 64 {
		return errors.New(utils.INVALID_INPUT)
	}

	if len([]rune(task.Description)) > 512 {
		return errors.New(utils.INVALID_INPUT)
	}

	return nil
}

func HandleCreateTask(w http.ResponseWriter, r *http.Request) utils.HandlerResponse {
	uid, err := firebase.GetFirebaseAuth().Verify(r.Header.Get("Authorization"))
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.AUTHORIZATION_ERROR), err)
	}

	input := Task{}
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.DECODER_ERROR), err)
	}

	err = parseTask(input)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.DECODER_ERROR), err)
	}

	err = createTask(db.GetSQLClient(), input, uid)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	return utils.MakeHandlerResponse(http.StatusOK, struct{}{}, nil)
}

func getTaskCustomer(client *db.SQLClient, taskID utils.UID) (utils.UID, error) {
	customer := utils.UID(0)
	reader, err := client.Query("SELECT customer_id FROM tasks WHERE task_id = $1", taskID)
	if err != nil {
		return customer, err
	}
	defer reader.Close()

	err = reader.GetRow(&customer)
	return customer, err
}

func setTaskDoer(client *db.SQLClient, taskID utils.UID, doerID utils.UID) error {
	reader, err := client.Query("UPDATE tasks SET doer_id = $2 WHERE task_id = $1", taskID, doerID)
	reader.Close()
	return err
}

func canSetDoer(userID utils.UID, customerID utils.UID) bool {
	return userID == customerID
}

func HandleSetDoer(w http.ResponseWriter, r *http.Request) utils.HandlerResponse {
	uid, err := firebase.GetFirebaseAuth().Verify(r.Header.Get("Authorization"))
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.AUTHORIZATION_ERROR), err)
	}

	doer := utils.UserData{}
	err = json.NewDecoder(r.Body).Decode(&doer)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.DECODER_ERROR), err)
	}

	taskID := utils.UID(0)
	err = taskID.FromString(mux.Vars(r)["task"])
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.DECODER_ERROR), err)
	}

	customer, err := getTaskCustomer(db.GetSQLClient(), taskID)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	if !canSetDoer(uid, customer) {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.AUTHORIZATION_ERROR), errors.New(utils.INSUFFICIENT_RIGHTS))
	}

	err = setTaskDoer(db.GetSQLClient(), taskID, doer.ID)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	return utils.MakeHandlerResponse(http.StatusOK, struct{}{}, nil)

}
