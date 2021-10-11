package db

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type PostTagLink struct {
	PostID int
	TagID  int
}

type PostTagLinkReader interface {
	Next(*PostTagLink) (bool, error)
}

type SQLPostTagLinkReader struct {
	rows *sql.Rows
}

func (fetcher *SQLPostTagLinkReader) Next(data *PostTagLink) (bool, error) {
	if fetcher.rows.Next() {
		err := fetcher.rows.Scan(&data.PostID, &data.TagID)
		return err == nil, err
	}
	return false, fetcher.rows.Err()
}

type ProfilesFetcher interface {
	GetUserProfile(id string) (PostTagLinkReader, error)
	GetPostsTags(id string) (PostTagLinkReader, error)
}

type SQLClient struct {
	db *sql.DB
}

func (client *SQLClient) GetUserProfile(id string) (PostTagLinkReader, error) {
	rows, err := client.db.Query("SELECT likes.post_id, tag_id FROM likes JOIN hashtags ON likes.post_id = hashtags.post_id WHERE user_id = $1", id)
	if err != nil {
		return nil, err
	}

	return &SQLPostTagLinkReader{rows: rows}, nil
}

func (client *SQLClient) GetPostsTags(id string) (PostTagLinkReader, error) {
	rows, err := client.db.Query("SELECT hashtags.post_id, tag_id FROM likes RIGHT JOIN hashtags ON likes.post_id = hashtags.post_id AND user_id = $1 WHERE user_id IS NULL", id)
	if err != nil {
		return nil, err
	}

	return &SQLPostTagLinkReader{rows: rows}, nil
}

var client *SQLClient

func GetSQLClient() *SQLClient {
	return client
}

func OpenDB(connectionString string) error {
	d, err := sql.Open("postgres", connectionString)
	if err == nil {
		client = &SQLClient{db: d}
		err = client.db.Ping()
	}
	return err
}
