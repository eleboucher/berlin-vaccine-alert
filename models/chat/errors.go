package chat

import "errors"

var (
	// ErrChatAlreadyExist is return when the chat already exist
	ErrChatAlreadyExist = errors.New("chat already exist")

	// ErrChatNotFound is return when the chat is not found
	ErrChatNotFound = errors.New("chat not found")
)
