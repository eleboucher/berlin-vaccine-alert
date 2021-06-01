package chat

import (
	"database/sql"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
)

var (
	tableName = "chats"

	fields = []string{
		"id",
	}

	preparedFields = strings.Join(fields, ", ")
)

// Chat holds the information for a telegram chat
type Chat struct {
	ID int64
}

// Model holds the information for the model
type Model struct {
	db *sql.DB
}

// NewModel returns a new model
func NewModel(db *sql.DB) *Model {
	return &Model{db: db}
}

// getSelectBuilder returns a SELECT statement builder for the chat model
func (model *Model) getSelectBuilder() sq.SelectBuilder {
	return sq.
		Select(fields...).
		From(tableName).
		RunWith(model.db)
}

// getInsertBuilder returns a INSERT statement builder for the chat model
func (model *Model) getInsertBuilder() sq.InsertBuilder {
	return sq.
		Insert(tableName).
		RunWith(model.db).
		Suffix(fmt.Sprintf("RETURNING %s", preparedFields))
}

// getUpdateBuilder returns a Update statement builder for the chat model
func (model *Model) getUpdateBuilder() sq.UpdateBuilder {
	return sq.
		Update(tableName).
		RunWith(model.db).
		Suffix(fmt.Sprintf("RETURNING %s", preparedFields))
}
