package chat

// Create creates a chat
func (m *Model) Create(id int64) (*Chat, error) {

	row := m.getInsertBuilder().Columns("id").Values(id).RunWith(m.db).QueryRow()
	chat, err := scanRow(row)
	if err != nil {
		if err.Error() == "UNIQUE constraint failed: chats.id" {
			return nil, ErrChatAlreadyExist
		}
		return nil, err
	}

	return chat, nil
}
