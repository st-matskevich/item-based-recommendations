package db

import (
	"database/sql"

	_ "github.com/lib/pq"
)

//TODO: now exposing only necessary methods, but maybe there is no need in encapsulation?
type SQLClient struct {
	db *sql.DB
}

var client *SQLClient

func (client *SQLClient) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return client.db.Query(query, args...)
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
