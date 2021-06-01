package chat

import (
	"database/sql"
	"strings"

	sq "github.com/Masterminds/squirrel"
)

// UpdateFilters update filters for the chat
func (m *Model) UpdateFilters(id int64, filter string) (*Chat, error) {
	var newFilters *string
	chat, err := m.Find(id)
	if err != nil {
		return nil, err
	}
	if isAlreadyFilter(chat.Filters, filter) {
		return chat, nil
	}

	if filter != "" {
		chat.Filters = append(chat.Filters, filter)
		tmp := strings.Join(chat.Filters, ",")
		newFilters = &tmp
	}

	row := m.getUpdateBuilder().Where(sq.Eq{"id": id}).Set("filters", newFilters).RunWith(m.db).QueryRow()
	chat, err = scanRow(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrChatNotFound
		}

		return nil, err
	}

	return chat, nil
}

func isAlreadyFilter(filters []string, toAdd string) bool {
	for _, filter := range filters {
		if filter == toAdd {
			return true
		}
	}
	return false
}
