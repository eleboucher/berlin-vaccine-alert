-- +migrate Up
CREATE TABLE IF NOT EXISTS chats (id INTEGER PRIMARY KEY);


-- +migrate Down
DROP TABLE chats;
