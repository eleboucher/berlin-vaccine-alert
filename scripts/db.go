package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func NewDB() (*sql.DB, error) {
	database, err := sql.Open("sqlite3", "../database.db")
	if err != nil {
		return nil, err
	}
	return database, nil
}
