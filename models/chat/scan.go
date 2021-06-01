package chat

import (
	"database/sql"

	sq "github.com/Masterminds/squirrel"
)

func scanRow(scanner sq.RowScanner) (*Chat, error) {

	chat := &Chat{}

	err := scanner.Scan(
		&chat.ID,
	)
	if err != nil {
		return nil, err
	}

	return chat, nil
}

func scanRows(rows *sql.Rows) ([]*Chat, error) {
	chats := make([]*Chat, 0)

	for rows.Next() {
		deployment, err := scanRow(rows)
		if err != nil {
			return nil, err
		}
		chats = append(chats, deployment)
	}

	return chats, nil
}
