package chat

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

// List lists chats
func (m *Model) List(vaccineName string) ([]*Chat, error) {
	fmt.Println("vaccinename:", vaccineName)
	q := m.getSelectBuilder().Where(
		sq.Eq{"enabled": true}).
		Where(
			sq.Or{sq.Like{"filters": "%" + vaccineName + "%"}, sq.Eq{"filters": nil}},
		)
	fmt.Println(q.ToSql())

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
