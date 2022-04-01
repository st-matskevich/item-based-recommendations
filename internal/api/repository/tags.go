package repository

import (
	"github.com/st-matskevich/item-based-recommendations/internal/api/utils"
	"github.com/st-matskevich/item-based-recommendations/internal/db"
)

type Tag struct {
	ID   utils.UID `json:"id"`
	Text string    `json:"name"`
}

type TagsRepository interface {
	GetTaskTags(taskID utils.UID) ([]Tag, error)
	SearchTags(request string) ([]Tag, error)
}

type TagsSQLRepository struct {
	SQLClient *db.SQLClient
}

func (repo *TagsSQLRepository) GetTaskTags(taskID utils.UID) ([]Tag, error) {
	reader, err := repo.SQLClient.Query(
		`SELECT tags.tag_id, tags.text
		FROM task_tag JOIN tags 
		ON task_tag.tag_id = tags.tag_id
		AND task_tag.task_id = $1`, taskID,
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

func (repo *TagsSQLRepository) SearchTags(request string) ([]Tag, error) {
	reader, err := repo.SQLClient.Query(
		`SELECT tags.tag_id, tags.text
		FROM tags 
		WHERE tags.text LIKE CONCAT($1::text, '%') LIMIT 10`, request,
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
