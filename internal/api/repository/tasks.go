package repository

import (
	"time"

	"github.com/lib/pq"
	"github.com/st-matskevich/item-based-recommendations/internal/api/utils"
	"github.com/st-matskevich/item-based-recommendations/internal/db"
)

const (
	NOT_ASSIGNED_TASKS = "NOT_ASSIGNED"
	CUSTOMER_TASKS     = "CUSTOMER"
	DOER_TASKS         = "DOER"
	LIKED              = "LIKED"
	RECOMMENDATIONS    = "RECOMMENDATIONS"
	REPLIED            = "REPLIED"
)

type Task struct {
	ID           utils.UID        `json:"id"`
	Name         string           `json:"name"`
	Description  string           `json:"description,omitempty"`
	Customer     UserData         `json:"customer"`
	Closed       bool             `json:"closed"`
	Owns         bool             `json:"owns"`
	Liked        bool             `json:"liked"`
	RepliesCount int32            `json:"replies"`
	Tags         utils.JSONObject `json:"tags"`
	CreatedAt    time.Time        `json:"createdAt"`
}

type TasksRepository interface {
	GetTasksFeed(scope string, request string, userID utils.UID) ([]Task, error)
	GetTask(userID utils.UID, taskID utils.UID) (*Task, error)
	GetTasks(userID utils.UID, tasksID []utils.UID) ([]Task, error)
	GetTaskCustomer(taskID utils.UID) (utils.UID, error)
	SetTaskLike(userID utils.UID, taskID utils.UID, value bool) error
	CreateTask(task Task) (utils.UID, error)
	CloseTask(taskID utils.UID, doerID utils.UID) error
}

type TasksSQLRepository struct {
	SQLClient *db.SQLClient
}

func (repo *TasksSQLRepository) buildTaskQuery(filter string) string {
	return `SELECT tasks.task_id, tasks.name, tasks.description, tasks.doer_id IS NOT NULL AS closed, tasks.customer_id = $1 AS owns, likes.active IS NOT NULL AND likes.active AS liked, COUNT(DISTINCT replies.reply_id), JSON_AGG(DISTINCT JSONB_BUILD_OBJECT('id', ENCODE(tags.tag_id::text::bytea, 'base64'), 'text', tags.text)), users.user_id, users.name, tasks.created_at
		FROM tasks 
		JOIN users 
		ON tasks.customer_id = users.user_id
		LEFT JOIN likes
		ON likes.task_id = tasks.task_id
		AND likes.user_id = $1
		LEFT JOIN replies
		ON replies.task_id = tasks.task_id
		LEFT JOIN task_tag
		ON task_tag.task_id = tasks.task_id
		LEFT JOIN tags
		ON task_tag.tag_id = tags.tag_id ` +
		filter +
		` GROUP BY tasks.task_id, users.user_id, likes.like_id
		ORDER BY tasks.task_id DESC`
}

func (repo *TasksSQLRepository) getTasksFeedReader(scope string, request string, userID utils.UID) (db.ResponseReader, error) {
	switch scope {
	case CUSTOMER_TASKS:
		return repo.SQLClient.Query(repo.buildTaskQuery("WHERE tasks.customer_id = $1 AND tasks.name LIKE '%' || $2 || '%'"), userID, request)
	case DOER_TASKS:
		return repo.SQLClient.Query(repo.buildTaskQuery("WHERE tasks.doer_id = $1 AND tasks.name LIKE '%' || $2 || '%'"), userID, request)
	case LIKED:
		return repo.SQLClient.Query(repo.buildTaskQuery("WHERE likes.active IS NOT NULL AND likes.active AND tasks.name LIKE '%' || $2 || '%'"), userID, request)
	case REPLIED:
		return repo.SQLClient.Query(repo.buildTaskQuery("WHERE replies.creator_id = $1 AND tasks.name LIKE '%' || $2 || '%'"), userID, request)
	}
	return repo.SQLClient.Query(repo.buildTaskQuery("WHERE tasks.doer_id IS NULL AND tasks.name LIKE '%' || $2 || '%'"), userID, request)
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
		ok, err := reader.NextRow(&row.ID, &row.Name, &row.Description, &row.Closed, &row.Owns, &row.Liked, &row.RepliesCount, &row.Tags, &row.Customer.ID, &row.Customer.Name, &row.CreatedAt)
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

func (repo *TasksSQLRepository) GetTask(userID utils.UID, taskID utils.UID) (*Task, error) {
	reader, err := repo.SQLClient.Query(repo.buildTaskQuery("WHERE tasks.task_id = $2"), userID, taskID)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	result := Task{}
	err = reader.GetRow(&result.ID, &result.Name, &result.Description, &result.Closed, &result.Owns, &result.Liked, &result.RepliesCount, &result.Tags, &result.Customer.ID, &result.Customer.Name, &result.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (repo *TasksSQLRepository) GetTasks(userID utils.UID, tasksID []utils.UID) ([]Task, error) {
	reader, err := repo.SQLClient.Query(repo.buildTaskQuery("WHERE tasks.task_id = ANY($2)"), userID, pq.Array(tasksID))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	result := []Task{}
	row := Task{}

	for {
		ok, err := reader.NextRow(&row.ID, &row.Name, &row.Description, &row.Closed, &row.Owns, &row.Liked, &row.RepliesCount, &row.Tags, &row.Customer.ID, &row.Customer.Name, &row.CreatedAt)
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

func (repo *TasksSQLRepository) GetTaskCustomer(taskID utils.UID) (utils.UID, error) {
	reader, err := repo.SQLClient.Query("SELECT tasks.customer_id FROM tasks WHERE tasks.task_id = $1", taskID)
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
