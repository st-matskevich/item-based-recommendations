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
	Id          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Customer    string    `json:"customer"`
	CreatedAt   time.Time `json:"createdAt"`
}

func getTasksReader(client *db.SQLClient, id int64, scope string) (db.ResponseReader, error) {
	switch scope {
	case CUSTOMER:
		return client.Query(`SELECT post_id, posts.name, description, users.name, created_at 
				FROM posts JOIN users 
				ON customer_id = user_id
				WHERE customer_id = $1
				ORDER BY created_at`, id)
	case DOER:
		return client.Query(`SELECT post_id, posts.name, description, users.name, created_at  
				FROM posts JOIN users 
				ON customer_id = user_id
				WHERE doer_id = $1
				ORDER BY created_at`, id)
	}
	return client.Query(`SELECT post_id, posts.name, description, users.name, created_at  
			FROM posts JOIN users 
			ON customer_id = user_id
			WHERE doer_id IS NULL
			ORDER BY created_at`)
}

func getTasks(reader db.ResponseReader) ([]Task, error) {
	result := []Task{}
	row := Task{}
	ok, err := reader.Next(&row.Id, &row.Name, &row.Description, &row.Customer, &row.CreatedAt)
	for ; ok; ok, err = reader.Next(&row.Id, &row.Name, &row.Description, &row.Customer, &row.CreatedAt) {
		result = append(result, row)
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func HandleGetTasks(w http.ResponseWriter, r *http.Request) utils.HandlerResponse {
	uid, err := firebase.GetFirebaseAuth().Verify(r.Header.Get("Authorization"))
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusBadRequest, utils.MakeErrorMessage(utils.AUTHORIZATION_ERROR), err)
	}

	scope := mux.Vars(r)["scope"]
	reader, err := getTasksReader(db.GetSQLClient(), uid, scope)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}
	defer reader.Close()

	tasks, err := getTasks(reader)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	return utils.MakeHandlerResponse(http.StatusOK, tasks, nil)
}

func createTask(client *db.SQLClient, id int64, task Task) error {
	reader, err := client.Query("INSERT INTO posts(name, description, customer_id) VALUES ($1, $2, $3)", task.Name, task.Description, id)
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

	err = createTask(db.GetSQLClient(), uid, input)
	if err != nil {
		return utils.MakeHandlerResponse(http.StatusInternalServerError, utils.MakeErrorMessage(utils.SQL_ERROR), err)
	}

	return utils.MakeHandlerResponse(http.StatusOK, struct{}{}, nil)
}
