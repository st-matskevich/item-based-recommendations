package db

import (
	"database/sql"
)

type ResponseReader interface {
	Next(dest ...interface{}) (bool, error)
	Close()
}

type SQLResponseReader struct {
	rows *sql.Rows
}

func (reader *SQLResponseReader) Next(dest ...interface{}) (bool, error) {
	if reader.rows.Next() {
		err := reader.rows.Scan(dest...)
		return err == nil, err
	}
	return false, reader.rows.Err()
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
