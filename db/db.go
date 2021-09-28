package db

import (
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/st-matskevich/item-based-recommendations/db/model"
)

var db *sql.DB

type SQLPostTagLinkFetcher struct {
	rows *sql.Rows
}

func (fetcher *SQLPostTagLinkFetcher) Next(data *model.PostTagLink) (bool, error) {
	if fetcher.rows.Next() {
		err := fetcher.rows.Scan(&data.PostID, &data.TagID)
		return err == nil, err
	}
	return false, fetcher.rows.Err()
}

func OpenDB(connectionString string) error {
	d, err := sql.Open("postgres", connectionString)
	if err == nil {
		db = d
		err = db.Ping()
	}
	return err
}

func GetUserProfile(id int) (*SQLPostTagLinkFetcher, error) {
	rows, err := db.Query("SELECT likes.post_id, tag_id FROM likes JOIN hashtags ON likes.post_id = hashtags.post_id WHERE user_id = $1", id)
	if err != nil {
		return nil, err
	}

	return &SQLPostTagLinkFetcher{rows: rows}, nil
}

func GetPostsTags(id int) (*SQLPostTagLinkFetcher, error) {
	rows, err := db.Query("SELECT hashtags.post_id, tag_id FROM likes RIGHT JOIN hashtags ON likes.post_id = hashtags.post_id AND user_id = $1 WHERE user_id IS NULL", id)
	if err != nil {
		return nil, err
	}

	return &SQLPostTagLinkFetcher{rows: rows}, nil
}
