package repository

import (
	"time"

	"github.com/st-matskevich/item-based-recommendations/internal/api/utils"
	"github.com/st-matskevich/item-based-recommendations/internal/db"
)

const (
	NOT_ASSIGNED_TASKS = "NOT_ASSIGNED"
	CUSTOMER_TASKS     = "CUSTOMER"
	DOER_TASKS         = "DOER"
	LIKED              = "LIKED"
)

type Task struct {
	ID          utils.UID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Customer    UserData  `json:"customer"`
	Closed      bool      `json:"closed"`
	CreatedAt   time.Time `json:"createdAt"`
}

type TasksRepository interface {
	GetTasksFeed(scope string, request string, userID utils.UID) ([]Task, error)
	GetTask(taskID utils.UID) (*Task, error)
	IsTaskLiked(userID utils.UID, taskID utils.UID) (bool, error)
	SetTaskLike(userID utils.UID, taskID utils.UID, value bool) error
	CreateTask(task Task) (utils.UID, error)
	CloseTask(taskID utils.UID, doerID utils.UID) error
}

type TasksSQLRepository struct {
	SQLClient *db.SQLClient
}

func (repo *TasksSQLRepository) getTasksFeedReader(scope string, request string, userID utils.UID) (db.ResponseReader, error) {
	switch scope {
	case CUSTOMER_TASKS:
		return repo.SQLClient.Query(
			`SELECT tasks.task_id, tasks.name, tasks.doer_id IS NOT NULL, users.user_id, users.name, tasks.created_at
			FROM tasks 
			JOIN users 
			ON tasks.customer_id = users.user_id
			AND tasks.customer_id = $1
			AND tasks.name LIKE '%' || $2 || '%'
			ORDER BY tasks.task_id DESC`, userID, request,
		)
	case DOER_TASKS:
		return repo.SQLClient.Query(
			`SELECT tasks.task_id, tasks.name, tasks.doer_id IS NOT NULL, users.user_id, users.name, tasks.created_at
			FROM tasks 
			JOIN users 
			ON tasks.customer_id = users.user_id
			AND tasks.doer_id = $1
			AND tasks.name LIKE '%' || $2 || '%'
			ORDER BY tasks.task_id DESC`, userID, request,
		)
	case LIKED:
		return repo.SQLClient.Query(
			`SELECT tasks.task_id, tasks.name, tasks.doer_id IS NOT NULL, users.user_id, users.name, tasks.created_at
			FROM tasks 
			JOIN users 
			ON tasks.customer_id = users.user_id
			JOIN likes
			ON tasks.task_id = likes.task_id
			AND likes.user_id = $1
			AND likes.active = true
			AND tasks.name LIKE '%' || $2 || '%'
			ORDER BY tasks.task_id DESC`, userID, request,
		)
	}
	return repo.SQLClient.Query(
		`SELECT tasks.task_id, tasks.name, tasks.doer_id IS NOT NULL, users.user_id, users.name, tasks.created_at
		FROM tasks 
		JOIN users 
		ON tasks.customer_id = users.user_id
		AND tasks.doer_id IS NULL
		AND tasks.name LIKE '%' || $1 || '%'
		ORDER BY tasks.task_id DESC`, request,
	)
}

func (repo *TasksSQLRepository) GetTasksFeed(scope string, request string, userID utils.UID) ([]Task, error) {
	reader, err := repo.getTasksFeedReader(scope, request, userID)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	result := []Task{}
	row := Task{}

	for {
		ok, err := reader.NextRow(&row.ID, &row.Name, &row.Closed, &row.Customer.ID, &row.Customer.Name, &row.CreatedAt)
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}

		result = append(result, row)
	}

	return result, nil
}

func (repo *TasksSQLRepository) GetTask(taskID utils.UID) (*Task, error) {
	reader, err := repo.SQLClient.Query(
		`SELECT tasks.task_id, tasks.name, tasks.description, tasks.doer_id IS NOT NULL, users.user_id, users.name, tasks.created_at  
		FROM tasks 
		JOIN users 
		ON tasks.customer_id = users.user_id
		AND tasks.task_id = $1`, taskID,
	)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	result := Task{}
	err = reader.GetRow(&result.ID, &result.Name, &result.Description, &result.Closed, &result.Customer.ID, &result.Customer.Name, &result.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (repo *TasksSQLRepository) IsTaskLiked(userID utils.UID, taskID utils.UID) (bool, error) {
	reader, err := repo.SQLClient.Query(
		`SELECT likes.active 
		FROM likes 
		WHERE likes.task_id = $1
		AND likes.user_id = $2`, taskID, userID,
	)
	if err != nil {
		return false, err
	}
	defer reader.Close()

	active := false
	found, err := reader.NextRow(&active)
	if err != nil {
		return false, err
	}

	return found && active, nil
}

func (repo *TasksSQLRepository) SetTaskLike(userID utils.UID, taskID utils.UID, value bool) error {
	return repo.SQLClient.Exec(
		`INSERT INTO likes(user_id, task_id, active) VALUES ($1, $2, $3)
		ON CONFLICT ON CONSTRAINT likes_user_task DO UPDATE SET active = $3`, userID, taskID, value,
	)
}

func (repo *TasksSQLRepository) CreateTask(task Task) (utils.UID, error) {
	reader, err := repo.SQLClient.Query("INSERT INTO tasks(name, description, customer_id) VALUES ($1, $2, $3) RETURNING task_id", task.Name, task.Description, task.Customer.ID)
	if err != nil {
		return 0, err
	}
	defer reader.Close()

	row := utils.UID(0)
	err = reader.GetRow(&row)
	if err != nil {
		return 0, err
	}

	return row, nil
}

func (repo *TasksSQLRepository) CloseTask(taskID utils.UID, doerID utils.UID) error {
	return repo.SQLClient.Exec("UPDATE tasks SET doer_id = $2 WHERE task_id = $1", taskID, doerID)
}
