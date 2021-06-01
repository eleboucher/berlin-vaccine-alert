package chat

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
)

// Delete a chat
func (m *Model) Delete(id int64) (*Chat, error) {

	row := m.getUpdateBuilder().Where(sq.Eq{"id": id}).Set("enabled", false).RunWith(m.db).QueryRow()
	chat, err := scanRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrChatNotFound
		}

		return nil, err
	}

	return chat, nil
}
