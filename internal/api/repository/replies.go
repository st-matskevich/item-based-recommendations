package repository

import (
	"time"

	"github.com/st-matskevich/item-based-recommendations/internal/api/utils"
	"github.com/st-matskevich/item-based-recommendations/internal/db"
)

type Reply struct {
	ID        utils.UID `json:"id"`
	Text      string    `json:"text"`
	Creator   UserData  `json:"creator"`
	CreatedAt time.Time `json:"createdAt"`
}

type RepliesRepository interface {
	GetReplies(taskID utils.UID) ([]Reply, error)
	GetDoerReply(taskID utils.UID) (*Reply, error)
	GetUserReply(taskID utils.UID, userID utils.UID) (*Reply, error)
	GetReply(replyID utils.UID) (*Reply, error)
	CreateReply(taskID utils.UID, reply Reply) (utils.UID, error)
	HideReply(replyID utils.UID) error
}

type RepliesSQLRepository struct {
	SQLClient *db.SQLClient
}

func (repo *RepliesSQLRepository) GetReplies(taskID utils.UID) ([]Reply, error) {
	reader, err := repo.SQLClient.Query(
		`SELECT replies.reply_id, replies.text, users.user_id, users.name, replies.created_at
		FROM tasks JOIN replies 
		ON tasks.task_id = replies.task_id
		AND tasks.task_id = $1
		AND replies.hidden = false
		JOIN users
		ON replies.creator_id = users.user_id
		ORDER BY replies.reply_id`, taskID,
	)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	replies := []Reply{}
	row := Reply{}
	for {
		ok, err := reader.NextRow(&row.ID, &row.Text, &row.Creator.ID, &row.Creator.Name, &row.CreatedAt)
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}

		replies = append(replies, row)
	}

	return replies, nil
}

func (repo *RepliesSQLRepository) GetDoerReply(taskID utils.UID) (*Reply, error) {
	reader, err := repo.SQLClient.Query(
		`SELECT replies.reply_id, replies.text, users.user_id, users.name, replies.created_at
		FROM tasks JOIN replies 
		ON tasks.task_id = replies.task_id
		AND tasks.task_id = $1
		AND tasks.doer_id = replies.creator_id
		JOIN users
		ON replies.creator_id = users.user_id`, taskID,
	)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	row := Reply{}
	found, err := reader.NextRow(&row.ID, &row.Text, &row.Creator.ID, &row.Creator.Name, &row.CreatedAt)
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, nil
	}

	return &row, nil
}

func (repo *RepliesSQLRepository) GetUserReply(taskID utils.UID, userID utils.UID) (*Reply, error) {
	reader, err := repo.SQLClient.Query(
		`SELECT replies.reply_id, replies.text, users.user_id, users.name, replies.created_at
		FROM tasks JOIN replies 
		ON tasks.task_id = replies.task_id
		AND tasks.task_id = $1
		AND replies.creator_id = $2
		JOIN users
		ON replies.creator_id = users.user_id`, taskID, userID,
	)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	row := Reply{}
	found, err := reader.NextRow(&row.ID, &row.Text, &row.Creator.ID, &row.Creator.Name, &row.CreatedAt)
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, nil
	}

	return &row, nil
}

func (repo *RepliesSQLRepository) CreateReply(taskID utils.UID, reply Reply) (utils.UID, error) {
	reader, err := repo.SQLClient.Query("INSERT INTO replies(task_id, text, creator_id) VALUES ($1, $2, $3) RETURNING reply_id", taskID, reply.Text, reply.Creator.ID)
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

func (repo *RepliesSQLRepository) HideReply(replyID utils.UID) error {
	return repo.SQLClient.Exec("UPDATE replies SET hidden = TRUE WHERE reply_id = $1", replyID)
}

func (repo *RepliesSQLRepository) GetReply(replyID utils.UID) (*Reply, error) {
	reader, err := repo.SQLClient.Query(
		`SELECT replies.reply_id, replies.text, users.user_id, users.name, replies.created_at
		FROM replies 
		JOIN users
		ON replies.creator_id = users.user_id
		AND replies.reply_id = $1`, replyID,
	)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	row := Reply{}
	err = reader.GetRow(&row.ID, &row.Text, &row.Creator.ID, &row.Creator.Name, &row.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &row, nil
}
