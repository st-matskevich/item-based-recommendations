package repository

import (
	"github.com/st-matskevich/item-based-recommendations/internal/api/utils"
	"github.com/st-matskevich/item-based-recommendations/internal/db"
)

type Tag struct {
	ID   utils.UID `json:"id"`
	Text string    `json:"text"`
}

type TaskTagLink struct {
	TaskID utils.UID
	TagID  utils.UID
}

type TagsRepository interface {
	SearchTags(request string) ([]Tag, error)
	CreateTag(tag string) (utils.UID, error)
	AddTagToTask(taskID utils.UID, tagID utils.UID) error
	GetTasksTags(userID utils.UID) ([]TaskTagLink, error)
}

type TagsSQLRepository struct {
	SQLClient *db.SQLClient
}

func (repo *TagsSQLRepository) SearchTags(request string) ([]Tag, error) {
	reader, err := repo.SQLClient.Query(
		`SELECT tags.tag_id, tags.text
		FROM tags 
		WHERE tags.text LIKE $1 || '%' LIMIT 10`, request,
	)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	tags := []Tag{}
	row := Tag{}
	for {
		ok, err := reader.NextRow(&row.ID, &row.Text)
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}

		tags = append(tags, row)
	}

	return tags, nil
}

func (repo *TagsSQLRepository) CreateTag(tag string) (utils.UID, error) {
	reader, err := repo.SQLClient.Query(
		`WITH new_tag AS (
			INSERT INTO tags(text)
			VALUES ($1)
			ON CONFLICT (text) DO NOTHING
			RETURNING tag_id
		) SELECT COALESCE(
			(SELECT tag_id FROM new_tag),
			(SELECT tag_id FROM tags WHERE text = $1)
		)`, tag)
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

func (repo *TagsSQLRepository) AddTagToTask(taskID utils.UID, tagID utils.UID) error {
	return repo.SQLClient.Exec("INSERT INTO task_tag(task_id, tag_id) VALUES ($1, $2)", taskID, tagID)
}

func (repo *TagsSQLRepository) GetTasksTags(userID utils.UID) ([]TaskTagLink, error) {
	reader, err := repo.SQLClient.Query(
		`SELECT task_tag.task_id, task_tag.tag_id 
		FROM likes 
		RIGHT JOIN task_tag 
		ON likes.task_id = task_tag.task_id 
		AND likes.user_id = $1 AND likes.active = true 
		WHERE likes.user_id IS NULL`, userID,
	)

	if err != nil {
		return nil, err
	}
	defer reader.Close()

	result := []TaskTagLink{}
	row := TaskTagLink{}
	for {
		ok, err := reader.NextRow(&row.TaskID, &row.TagID)
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
