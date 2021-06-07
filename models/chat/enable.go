package chat

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
)

// Enable a chat
func (m *Model) Enable(id int64) (*Chat, error) {

	row := m.getUpdateBuilder().Where(sq.Eq{"id": id}).Set("enabled", true).QueryRow()
	chat, err := scanRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrChatNotFound
		}

		return nil, err
	}

	return chat, nil
}
