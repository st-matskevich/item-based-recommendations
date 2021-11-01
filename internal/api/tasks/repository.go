package tasks

import (
	"time"

	"github.com/st-matskevich/item-based-recommendations/internal/api/profile"
	"github.com/st-matskevich/item-based-recommendations/internal/api/utils"
	"github.com/st-matskevich/item-based-recommendations/internal/db"
)

const (
	NOT_ASSIGNED = "NOT_ASSIGNED"
	CUSTOMER     = "CUSTOMER"
	DOER         = "DOER"
)

type Task struct {
	ID           utils.UID        `json:"id"`
	Name         string           `json:"name"`
	Description  string           `json:"description,omitempty"`
	Customer     profile.UserData `json:"customer"`
	Owns         bool             `json:"owns"`
	Closed       bool             `json:"closed"`
	RepliesCount int32            `json:"repliesCount"`
	CreatedAt    time.Time        `json:"createdAt"`
}

type TasksRepository interface {
	GetTasksFeed(scope string, userID utils.UID) ([]Task, error)
	GetTask(taskID utils.UID) (*Task, error)
	CreateTask(task Task) error
	CloseTask(taskID utils.UID, doerID utils.UID) error
}

type TasksSQLRepository struct {
	SQLClient *db.SQLClient
}

func (repo *TasksSQLRepository) getTasksFeedReader(scope string, userID utils.UID) (db.ResponseReader, error) {
	switch scope {
	case CUSTOMER:
		return repo.SQLClient.Query(
			`SELECT tasks.task_id, tasks.name, tasks.doer_id IS NOT NULL, users.user_id, users.name, COUNT(replies.task_id), tasks.created_at
			FROM tasks 
			JOIN users 
			ON tasks.customer_id = users.user_id
			AND tasks.customer_id = $1
			LEFT JOIN replies
			ON tasks.task_id = replies.task_id
			GROUP BY tasks.task_id, tasks.name, tasks.doer_id, users.user_id, users.name, tasks.created_at
			ORDER BY tasks.task_id`, userID,
		)
	case DOER:
		return repo.SQLClient.Query(
			`SELECT tasks.task_id, tasks.name, tasks.doer_id IS NOT NULL, users.user_id, users.name, COUNT(replies.task_id), tasks.created_at
			FROM tasks 
			JOIN users 
			ON tasks.customer_id = users.user_id
			AND tasks.doer_id = $1
			LEFT JOIN replies
			ON tasks.task_id = replies.task_id
			GROUP BY tasks.task_id, tasks.name, tasks.doer_id, users.user_id, users.name, tasks.created_at
			ORDER BY tasks.task_id`, userID,
		)
	}
	return repo.SQLClient.Query(
		`SELECT tasks.task_id, tasks.name, tasks.doer_id IS NOT NULL, users.user_id, users.name, COUNT(replies.task_id), tasks.created_at
		FROM tasks 
		JOIN users 
		ON tasks.customer_id = users.user_id
		AND tasks.doer_id IS NULL
		LEFT JOIN replies
		ON tasks.task_id = replies.task_id
		GROUP BY tasks.task_id, tasks.name, tasks.doer_id, users.user_id, users.name, tasks.created_at
		ORDER BY tasks.task_id`,
	)
}

func (repo *TasksSQLRepository) GetTasksFeed(scope string, userID utils.UID) ([]Task, error) {
	reader, err := repo.getTasksFeedReader(scope, userID)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	result := []Task{}
	row := Task{}

	for {
		ok, err := reader.NextRow(&row.ID, &row.Name, &row.Closed, &row.Customer.ID, &row.Customer.Name, &row.RepliesCount, &row.CreatedAt)
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
		`SELECT tasks.task_id, tasks.name, tasks.description, tasks.doer_id IS NOT NULL, users.user_id, users.name, COUNT(replies.task_id), tasks.created_at  
		FROM tasks 
		JOIN users 
		ON tasks.customer_id = users.user_id
		AND tasks.task_id = $1
		LEFT JOIN replies
		ON tasks.task_id = replies.task_id
		GROUP BY tasks.task_id, tasks.name, tasks.description, tasks.doer_id, users.user_id, users.name, tasks.created_at
		ORDER BY tasks.task_id`, taskID,
	)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	result := Task{}
	err = reader.GetRow(&result.ID, &result.Name, &result.Description, &result.Closed, &result.Customer.ID, &result.Customer.Name, &result.RepliesCount, &result.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (repo *TasksSQLRepository) CreateTask(task Task) error {
	return repo.SQLClient.Exec("INSERT INTO tasks(name, description, customer_id) VALUES ($1, $2, $3)", task.Name, task.Description, task.Customer.ID)
}

func (repo *TasksSQLRepository) CloseTask(taskID utils.UID, doerID utils.UID) error {
	return repo.SQLClient.Exec("UPDATE tasks SET doer_id = $2 WHERE task_id = $1", taskID, doerID)
}
