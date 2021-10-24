package tasks

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
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
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Customer    string    `json:"customer"`
	CreatedAt   time.Time `json:"createdAt"`
}

func getTasksFeedReader(client *db.SQLClient, userID int64, scope string) (db.ResponseReader, error) {
	switch scope {
	case CUSTOMER:
		return client.Query(`SELECT task_id, tasks.name, users.name, created_at 
							FROM tasks JOIN users 
							ON customer_id = user_id
							AND customer_id = $1
							ORDER BY created_at`, userID)
	case DOER:
		return client.Query(`SELECT task_id, tasks.name, users.name, created_at  
							FROM tasks JOIN users 
							ON customer_id = user_id
							AND doer_id = $1
							ORDER BY created_at`, userID)
	}
	return client.Query(`SELECT task_id, tasks.name, users.name, created_at  
						FROM tasks JOIN users 
						ON customer_id = user_id
						AND doer_id IS NULL
						ORDER BY created_at`)
}

func getTasksFeed(reader db.ResponseReader) ([]Task, error) {
	result := []Task{}
	row := Task{}
	ok, err := reader.Next(&row.ID, &row.Name, &row.Customer, &row.CreatedAt)
	for ; ok; ok, err = reader.Next(&row.ID, &row.Name, &row.Customer, &row.CreatedAt) {
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

	scope := mux.Vars(r)["scope"]
	reader, err := getTasksFeedReader(db.GetSQLClient(), uid, scope)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}
	defer reader.Close()

	tasks, err := getTasksFeed(reader)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	return utils.MakeHandlerResponse(http.StatusOK, tasks, nil)
}

func getTaskReader(client *db.SQLClient, taskID int64) (db.ResponseReader, error) {
	return client.Query(`SELECT task_id, tasks.name, description, users.name, created_at  
						FROM tasks JOIN users 
						ON customer_id = user_id
						AND task_id = $1`, taskID)
}

func getTask(reader db.ResponseReader) (Task, error) {
	result := Task{}
	found, err := reader.Next(&result.ID, &result.Name, &result.Description, &result.Customer, &result.CreatedAt)
	if !found && err == nil {
		err = errors.New(utils.SQL_NO_RESULT)
	}
	return result, err
}

func HandleGetTask(w http.ResponseWriter, r *http.Request) utils.HandlerResponse {
	_, err := firebase.GetFirebaseAuth().Verify(r.Header.Get("Authorization"))
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.AUTHORIZATION_ERROR), err)
	}

	taskID, err := strconv.ParseInt(mux.Vars(r)["task"], 10, 64)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.DECODER_ERROR), err)
	}

	reader, err := getTaskReader(db.GetSQLClient(), taskID)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}
	defer reader.Close()

	task, err := getTask(reader)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	return utils.MakeHandlerResponse(http.StatusOK, task, nil)
}

func createTask(client *db.SQLClient, task Task, userID int64) error {
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
