package chat

import sq "github.com/Masterminds/squirrel"

// List lists chats
func (m *Model) List() ([]*Chat, error) {

	q := m.getSelectBuilder().Where(sq.Eq{"enabled": true})
	rows, err := q.Query()
	if err != nil {
		return nil, err
	}

	chats, err := scanRows(rows)
	if err != nil {
		return nil, err
	}

	return chats, nil
}
