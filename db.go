package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func NewDB() (*sql.DB, error) {
	database, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		return nil, err
	}
	statement, err := database.Prepare("CREATE TABLE IF NOT EXISTS chats (id INTEGER PRIMARY KEY)")
	if err != nil {
		return nil, err
	}
	statement.Exec()
	return database, nil
}
