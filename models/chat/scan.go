package chat

import (
	"database/sql"
	"strings"

	sq "github.com/Masterminds/squirrel"
)

func scanRow(scanner sq.RowScanner) (*Chat, error) {
	var filters *string

	chat := &Chat{}
	err := scanner.Scan(
		&chat.ID,
		&filters,
	)

	if filters != nil {
		chat.Filters = strings.Split(*filters, ",")
	}
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
