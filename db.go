package main

import (
	"database/sql"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/spf13/viper"
)

func NewDB() (*sql.DB, error) {
	database, err := sql.Open("pgx", viper.GetString("DATABASE_URL"))
	if err != nil {
		return nil, err
	}
	return database, nil
}
