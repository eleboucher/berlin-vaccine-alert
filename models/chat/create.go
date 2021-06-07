package chat

import (
	"errors"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
)

// Create creates a chat
func (m *Model) Create(id int64) (*Chat, error) {

	row := m.getInsertBuilder().Columns("id").Values(id).QueryRow()
	chat, err := scanRow(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return nil, ErrChatAlreadyExist
			}
		}
		return nil, err

	}

	return chat, nil
}
