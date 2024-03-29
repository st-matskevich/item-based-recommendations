package db

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type ResponseReader interface {
	NextRow(dest ...interface{}) (bool, error)
	GetRow(dest ...interface{}) error
	Close()
}

type SQLResponseReader struct {
	rows *sql.Rows
}

func (reader *SQLResponseReader) NextRow(dest ...interface{}) (bool, error) {
	if reader.rows.Next() {
		err := reader.rows.Scan(dest...)
		return err == nil, err
	}
	return false, reader.rows.Err()
}

func (reader *SQLResponseReader) GetRow(dest ...interface{}) error {
	found, err := reader.NextRow(dest...)
	if !found && err == nil {
		err = sql.ErrNoRows
	}
	return err
}

func (reader *SQLResponseReader) Close() {
	if reader.rows != nil {
		reader.rows.Close()
	}
}

type SQLClient struct {
	db *sql.DB
}

var client *SQLClient

func (client *SQLClient) Query(query string, args ...interface{}) (*SQLResponseReader, error) {
	response, err := client.db.Query(query, args...)
	return &SQLResponseReader{response}, err
}

func (client *SQLClient) Exec(query string, args ...interface{}) error {
	_, err := client.db.Exec(query, args...)
	return err
}

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
