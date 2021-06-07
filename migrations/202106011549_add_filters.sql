-- +migrate Up
ALTER TABLE chats ADD COLUMN IF NOT EXISTS filters TEXT;


-- +migrate Down
ALTER TABLE chats DROP COLUMN filters;
