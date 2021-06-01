-- +migrate Up
ALTER TABLE chats ADD COLUMN 'enabled' BOOLEAN NOT NULL DEFAULT TRUE;


-- +migrate Down
ALTER TABLE chats DROP COLUMN 'enabled';
