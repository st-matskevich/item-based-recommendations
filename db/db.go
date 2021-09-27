//TODO: because of the for loops in GetPostsTags and GetUserProfile service will have worse performance,
//should rewrite this module and implement some mockable interfaces for the module "parser" with methods like rows.Next()
package db

import (
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/st-matskevich/item-based-recommendations/model"
)

var db *sql.DB

func OpenDB(connectionString string) error {
	d, err := sql.Open("postgres", connectionString)
	if err == nil {
		db = d
		err = db.Ping()
	}
	return err
}

func GetUserProfile(id int) ([]model.PostTagLink, error) {
	rows, err := db.Query("SELECT likes.post_id, tag_id FROM likes JOIN hashtags ON likes.post_id = hashtags.post_id WHERE user_id = $1", id)
	if err != nil {
		return nil, err
	}

	result := []model.PostTagLink{}
	for rows.Next() {
		row := model.PostTagLink{}
		err = rows.Scan(&row.PostID, &row.TagID)
		if err != nil {
			return nil, err
		}
		result = append(result, row)
	}

	return result, nil
}

func GetPostsTags(id int) ([]model.PostTagLink, error) {
	rows, err := db.Query("SELECT hashtags.post_id, tag_id FROM likes RIGHT JOIN hashtags ON likes.post_id = hashtags.post_id AND user_id = $1 WHERE user_id IS NULL", id)
	if err != nil {
		return nil, err
	}

	result := []model.PostTagLink{}
	for rows.Next() {
		row := model.PostTagLink{}
		err = rows.Scan(&row.PostID, &row.TagID)
		if err != nil {
			return nil, err
		}
		result = append(result, row)
	}

	return result, nil
}
