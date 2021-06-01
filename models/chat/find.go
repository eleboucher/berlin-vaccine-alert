package chat

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
)

// Find find a chat
func (m *Model) Find(id int64) (*Chat, error) {

	q := m.getSelectBuilder().Where(sq.Eq{"id": id}).Limit(1)

	chat, err := scanRow(q)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrChatNotFound
		}
		return nil, err
	}

	return chat, nil
}
