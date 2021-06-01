-- +migrate Up
ALTER TABLE chats ADD COLUMN 'filters' TEXT;


-- +migrate Down
ALTER TABLE chats DROP COLUMN 'filters';
